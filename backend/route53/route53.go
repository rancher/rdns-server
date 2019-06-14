package route53

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rancher/rdns-server/database"
	"github.com/rancher/rdns-server/model"
	"github.com/rancher/rdns-server/util"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	typeA            = "A"
	typeTXT          = "TXT"
	typeToken        = "TOKEN"
	typeFrozen       = "FROZEN"
	maxSlugHashTimes = 100
	slugLength       = 6
	tokenLength      = 32
	route53TTL       = 300
)

type Backend struct {
	TTL    time.Duration
	Zone   string
	ZoneID string

	Svc *route53.Route53
}

func NewBackend() (*Backend, error) {
	c := credentials.NewEnvCredentials()

	s, err := session.NewSession()
	if err != nil {
		return &Backend{}, err
	}

	svc := route53.New(s, &aws.Config{
		Credentials: c,
	})

	z, err := svc.GetHostedZone(&route53.GetHostedZoneInput{
		Id: aws.String(os.Getenv("AWS_HOSTED_ZONE_ID")),
	})
	if err != nil {
		return &Backend{}, err
	}

	d, err := time.ParseDuration(os.Getenv("TTL"))
	if err != nil {
		return &Backend{}, errors.Wrapf(err, errParseFlag, "ttl")
	}

	return &Backend{
		TTL:    d,
		Zone:   strings.TrimRight(aws.StringValue(z.HostedZone.Name), "."),
		ZoneID: aws.StringValue(z.HostedZone.Id),
		Svc:    svc,
	}, nil
}

func (b *Backend) GetZone() string {
	return b.Zone
}

func (b *Backend) Get(opts *model.DomainOptions) (d model.Domain, err error) {
	logrus.Debugf("get A record for domain options: %s", opts.String())

	records, err := b.getRecords(opts, typeA)
	if err != nil {
		return d, err
	}

	v, a, s, _ := b.filterRecords(records.ResourceRecordSets, opts, typeA)
	if !v {
		return d, errors.Errorf(errEmptyRecord, typeA, opts.String())
	}

	// convert A & sub domain records to map
	ca, cs := b.convertARecords(a, s)

	// get token from database
	token, err := database.GetDatabase().QueryToken(opts.Fqdn)
	if err != nil {
		return d, errors.Wrapf(err, errOperateDatabase, typeToken, opts.String())
	}

	d.Fqdn = opts.Fqdn
	d.Hosts = ca[opts.Fqdn]
	d.SubDomain = cs
	d.Expiration = convertExpiration(time.Unix(0, token.CreatedOn), int(b.TTL.Nanoseconds()))

	return d, nil
}

func (b *Backend) Set(opts *model.DomainOptions) (d model.Domain, err error) {
	logrus.Debugf("set A record for domain options: %s", opts.String())

	for i := 0; i < maxSlugHashTimes; i++ {
		fqdn := fmt.Sprintf("%s.%s", generateSlug(), b.Zone)

		// check whether this slug name can be used or not, if not found the slug name is valid others not valid
		r, err := database.GetDatabase().QueryFrozen(strings.Split(fqdn, ".")[0])
		if err != nil && err != sql.ErrNoRows {
			return d, err
		}
		if r != "" {
			logrus.Debugf(errNotValidGenerateName, strings.Split(fqdn, ".")[0])
			continue
		}

		o := &model.DomainOptions{
			Fqdn: fqdn,
		}

		d, err := b.Get(o)
		if err != nil || len(d.Hosts) == 0 {
			opts.Fqdn = fqdn
			break
		}
	}

	if opts.Fqdn == "" {
		return d, errors.Errorf(errGenerateName, opts.String())
	}

	// save the slug name to the database in case of the name will be re-generate
	if err := database.GetDatabase().InsertFrozen(strings.Split(opts.Fqdn, ".")[0]); err != nil {
		return d, errors.Wrapf(err, errOperateDatabase, typeToken, strings.Split(opts.Fqdn, ".")[0])
	}

	// save token to the database
	tID, err := b.SetToken(opts, false)
	if err != nil {
		return d, errors.Wrapf(err, errOperateDatabase, typeToken, opts.String())
	}

	// set A record
	rr := make([]*route53.ResourceRecord, 0)
	for _, h := range opts.Hosts {
		rr = append(rr, &route53.ResourceRecord{
			Value: aws.String(h),
		})
	}

	rrs := &route53.ResourceRecordSet{
		Type:            aws.String(typeA),
		Name:            aws.String(opts.Fqdn),
		ResourceRecords: rr,
		TTL:             aws.Int64(int64(route53TTL)),
	}
	pID, err := b.setRecord(rrs, opts, typeA, tID, 0, false)
	if err != nil {
		return d, err
	}

	// set wildcard A record
	rrs.Name = aws.String(fmt.Sprintf("\\052.%s", opts.Fqdn))
	if _, err := b.setRecord(rrs, opts, typeA, tID, pID, false); err != nil {
		return d, err
	}

	// set sub domain A record
	for k, v := range opts.SubDomain {
		rr := make([]*route53.ResourceRecord, 0)
		for _, h := range v {
			rr = append(rr, &route53.ResourceRecord{
				Value: aws.String(h),
			})
		}

		rrs := &route53.ResourceRecordSet{
			Type:            aws.String(typeA),
			Name:            aws.String(fmt.Sprintf("%s.%s", k, opts.Fqdn)),
			ResourceRecords: rr,
			TTL:             aws.Int64(int64(route53TTL)),
		}

		if _, err := b.setRecord(rrs, opts, typeA, tID, pID, true); err != nil {
			return d, err
		}
	}

	return b.Get(opts)
}

func (b *Backend) Update(opts *model.DomainOptions) (d model.Domain, err error) {
	logrus.Debugf("update A record for domain options: %s", opts.String())

	records, err := b.getRecords(opts, typeA)
	if err != nil {
		return d, err
	}

	v, a, s, _ := b.filterRecords(records.ResourceRecordSets, opts, typeA)
	if !v {
		return d, errors.Errorf(errEmptyRecord, typeA, opts.String())
	}

	// convert A & sub domain records to map
	_, cs := b.convertARecords(a, s)

	// update A records
	rr := make([]*route53.ResourceRecord, 0)
	for _, h := range opts.Hosts {
		rr = append(rr, &route53.ResourceRecord{
			Value: aws.String(h),
		})
	}

	rrs := &route53.ResourceRecordSet{
		Type:            aws.String(typeA),
		Name:            aws.String(opts.Fqdn),
		ResourceRecords: rr,
		TTL:             aws.Int64(int64(route53TTL)),
	}

	r, err := database.GetDatabase().QueryA(opts.Fqdn)
	if err != nil {
		return d, errors.Wrapf(err, errOperateDatabase, typeA, opts.String())
	}

	if _, err := b.setRecord(rrs, opts, typeA, r.TID, 0, false); err != nil {
		return d, err
	}

	// update wildcard A records
	rrs.Name = aws.String(fmt.Sprintf("\\052.%s", opts.Fqdn))
	if _, err := b.setRecord(rrs, opts, typeA, r.TID, r.ID, false); err != nil {
		return d, err
	}

	// update sub domain A records
	for k, v := range opts.SubDomain {
		rr := make([]*route53.ResourceRecord, 0)
		for _, h := range v {
			rr = append(rr, &route53.ResourceRecord{
				Value: aws.String(h),
			})
		}

		rrs := &route53.ResourceRecordSet{
			Type:            aws.String(typeA),
			Name:            aws.String(fmt.Sprintf("%s.%s", k, opts.Fqdn)),
			ResourceRecords: rr,
			TTL:             aws.Int64(int64(route53TTL)),
		}

		if _, err := b.setRecord(rrs, opts, typeA, r.TID, r.ID, true); err != nil {
			return d, err
		}

	}

	// delete useless sub domain A records
	for k, v := range cs {
		if _, ok := opts.SubDomain[k]; !ok {
			rr := make([]*route53.ResourceRecord, 0)
			for _, h := range v {
				rr = append(rr, &route53.ResourceRecord{
					Value: aws.String(h),
				})
			}

			rrs := &route53.ResourceRecordSet{
				Name:            aws.String(fmt.Sprintf("%s.%s", k, opts.Fqdn)),
				Type:            aws.String(typeA),
				ResourceRecords: rr,
				TTL:             aws.Int64(int64(route53TTL)),
			}

			if err := b.deleteRecord(rrs, opts, typeA, true); err != nil {
				return d, err
			}
			continue
		}
	}

	// get token from database
	token, err := database.GetDatabase().QueryToken(opts.Fqdn)
	if err != nil {
		return d, errors.Wrapf(err, errOperateDatabase, typeToken, opts.String())
	}

	d.Fqdn = opts.Fqdn
	d.Hosts = opts.Hosts
	d.SubDomain = opts.SubDomain
	d.Expiration = convertExpiration(time.Unix(0, token.CreatedOn), int(b.TTL.Nanoseconds()))

	return d, nil
}

func (b *Backend) Delete(opts *model.DomainOptions) error {
	logrus.Debugf("delete A record for domain options: %s", opts.String())

	records, err := b.getRecords(opts, typeA)
	if err != nil {
		return err
	}

	v, a, s, _ := b.filterRecords(records.ResourceRecordSets, opts, typeA)
	if !v {
		return errors.Errorf(errEmptyRecord, typeA, opts.String())
	}

	// delete A records and wildcard A records
	for _, rr := range a {
		if err := b.deleteRecord(rr, opts, typeA, false); err != nil {
			return err
		}
		rr.Name = aws.String(fmt.Sprintf("\\052.%s", opts.Fqdn))
		if err := b.deleteRecord(rr, opts, typeA, false); err != nil {
			return err
		}
	}

	// delete sub domain A records
	for _, rr := range s {
		if err := b.deleteRecord(rr, opts, typeA, true); err != nil {
			return err
		}
	}

	return nil
}

func (b *Backend) Renew(opts *model.DomainOptions) (d model.Domain, err error) {
	logrus.Debugf("renew records for domain options: %s", opts.String())

	// renew token record
	t, err := database.GetDatabase().QueryToken(opts.Fqdn)
	if err != nil {
		return d, errors.Wrapf(err, errOperateDatabase, typeToken, opts.String())
	}
	_, nt, err := database.GetDatabase().RenewToken(t.Token)
	if err != nil {
		return d, errors.Wrapf(err, errOperateDatabase, typeToken, opts.String())
	}

	// renew frozen record
	if err := database.GetDatabase().RenewFrozen(strings.Split(opts.Fqdn, ".")[0]); err != nil {
		return d, errors.Wrapf(err, errOperateDatabase, typeFrozen, opts.String())
	}

	d.Fqdn = opts.Fqdn
	d.Hosts = opts.Hosts
	d.SubDomain = opts.SubDomain
	d.Expiration = convertExpiration(time.Unix(0, nt), int(b.TTL.Nanoseconds()))

	return d, nil
}

func (b *Backend) GetText(opts *model.DomainOptions) (d model.Domain, err error) {
	logrus.Debugf("get TXT record for domain options: %s", opts.String())

	records, err := b.getRecords(opts, typeTXT)
	if err != nil {
		return d, err
	}

	valid, _, _, t := b.filterRecords(records.ResourceRecordSets, opts, typeTXT)
	if !valid || len(t) < 1 {
		return d, errors.Errorf(errEmptyRecord, typeTXT, opts.String())
	}

	// get token from database
	token, err := database.GetDatabase().QueryToken(b.findSlugWithZone(opts.Fqdn))
	if err != nil {
		return d, errors.Wrapf(err, errOperateDatabase, typeToken, opts.String())
	}

	d.Fqdn = opts.Fqdn
	d.Text = strings.Trim(aws.StringValue(t[0].ResourceRecords[0].Value), "\"")
	d.Expiration = convertExpiration(time.Unix(0, token.CreatedOn), int(b.TTL.Nanoseconds()))

	return d, nil
}

func (b *Backend) SetText(opts *model.DomainOptions) (d model.Domain, err error) {
	logrus.Debugf("set TXT record for domain options: %s", opts.String())

	records, err := b.getRecords(opts, typeTXT)
	if err != nil {
		return d, err
	}

	if valid, _, _, _ := b.filterRecords(records.ResourceRecordSets, opts, typeTXT); valid {
		return d, errors.Errorf(errExistRecord, typeTXT, opts.String())
	}

	r, err := database.GetDatabase().QueryToken(b.findSlugWithZone(opts.Fqdn))
	if err != nil {
		return d, errors.Wrapf(err, errOperateDatabase, typeTXT, opts.String())
	}

	rrs := &route53.ResourceRecordSet{
		Name: aws.String(opts.Fqdn),
		Type: aws.String(typeTXT),
		ResourceRecords: []*route53.ResourceRecord{
			{
				Value: aws.String(fmt.Sprintf("\"%s\"", opts.Text)),
			},
		},
		TTL: aws.Int64(int64(route53TTL)),
	}

	if _, err := b.setRecord(rrs, opts, typeTXT, r.ID, 0, false); err != nil {
		return d, err
	}

	return b.GetText(opts)
}

func (b *Backend) UpdateText(opts *model.DomainOptions) (d model.Domain, err error) {
	logrus.Debugf("update TXT record for domain options: %s", opts.String())

	records, err := b.getRecords(opts, typeTXT)
	if err != nil {
		return d, err
	}

	if valid, _, _, _ := b.filterRecords(records.ResourceRecordSets, opts, typeTXT); !valid {
		return d, errors.Errorf(errEmptyRecord, typeTXT, opts.String())
	}

	r, err := database.GetDatabase().QueryTXT(opts.Fqdn)
	if err != nil {
		return d, errors.Wrapf(err, errOperateDatabase, typeTXT, opts.String())
	}

	rrs := &route53.ResourceRecordSet{
		Name: aws.String(opts.Fqdn),
		Type: aws.String(typeTXT),
		ResourceRecords: []*route53.ResourceRecord{
			{
				Value: aws.String(fmt.Sprintf("\"%s\"", opts.Text)),
			},
		},
		TTL: aws.Int64(int64(route53TTL)),
	}

	if _, err := b.setRecord(rrs, opts, typeTXT, r.TID, 0, false); err != nil {
		return d, err
	}

	// get token from database
	token, err := database.GetDatabase().QueryToken(b.findSlugWithZone(opts.Fqdn))
	if err != nil {
		return d, errors.Wrapf(err, errOperateDatabase, typeToken, opts.String())
	}

	d.Fqdn = opts.Fqdn
	d.Hosts = opts.Hosts
	d.Text = opts.Text
	d.Expiration = convertExpiration(time.Unix(0, token.CreatedOn), int(b.TTL.Nanoseconds()))

	return d, nil
}

func (b *Backend) DeleteText(opts *model.DomainOptions) error {
	logrus.Debugf("delete TXT record for domain options: %s", opts.String())

	records, err := b.getRecords(opts, typeTXT)
	if err != nil {
		return err
	}

	v, _, _, t := b.filterRecords(records.ResourceRecordSets, opts, typeTXT)
	if !v {
		return errors.Errorf(errEmptyRecord, typeTXT, opts.String())
	}

	for _, rr := range t {
		if err := b.deleteRecord(rr, opts, typeTXT, false); err != nil {
			return err
		}
	}

	return nil
}

func (b *Backend) GetToken(fqdn string) (string, error) {
	t, err := database.GetDatabase().QueryToken(fqdn)
	return t.Token, err
}

func (b *Backend) SetToken(opts *model.DomainOptions, exist bool) (int64, error) {
	if exist {
		id, _, err := database.GetDatabase().RenewToken(opts.Fqdn)
		if err != nil {
			return 0, err
		}
		return id, err
	}

	return database.GetDatabase().InsertToken(generateToken(), opts.Fqdn)
}

func (b *Backend) MigrateFrozen(opts *model.MigrateFrozen) error {
	return database.GetDatabase().MigrateFrozen(opts.Path, opts.Expiration.UnixNano())
}

func (b *Backend) MigrateToken(opts *model.MigrateToken) error {
	return database.GetDatabase().MigrateToken(opts.Token, opts.Path, opts.Expiration.UnixNano())
}

func (b *Backend) MigrateRecord(opts *model.MigrateRecord) error {
	if opts.Text != "" {
		// migrate TXT record
		dopts := &model.DomainOptions{
			Fqdn: opts.Fqdn,
			Text: opts.Text,
		}
		if _, err := b.SetText(dopts); err != nil {
			return err
		}
	} else {
		// migrate A record
		dopts := &model.DomainOptions{
			Fqdn:      opts.Fqdn,
			Hosts:     opts.Hosts,
			SubDomain: opts.SubDomain,
		}
		t, err := database.GetDatabase().QueryToken(b.findSlugWithZone(dopts.Fqdn))
		if err != nil {
			return errors.Wrapf(err, errOperateDatabase, typeA, dopts.String())
		}

		rr := make([]*route53.ResourceRecord, 0)
		for _, h := range dopts.Hosts {
			rr = append(rr, &route53.ResourceRecord{
				Value: aws.String(h),
			})
		}

		rrs := &route53.ResourceRecordSet{
			Type:            aws.String(typeA),
			Name:            aws.String(dopts.Fqdn),
			ResourceRecords: rr,
			TTL:             aws.Int64(int64(route53TTL)),
		}
		pID, err := b.setRecord(rrs, dopts, typeA, t.ID, 0, false)
		if err != nil {
			return err
		}

		// set wildcard A record
		rrs.Name = aws.String(fmt.Sprintf("\\052.%s", dopts.Fqdn))
		if _, err := b.setRecord(rrs, dopts, typeA, t.ID, pID, false); err != nil {
			return err
		}

		// set sub domain A record
		for k, v := range dopts.SubDomain {
			rr := make([]*route53.ResourceRecord, 0)
			for _, h := range v {
				rr = append(rr, &route53.ResourceRecord{
					Value: aws.String(h),
				})
			}

			rrs := &route53.ResourceRecordSet{
				Type:            aws.String(typeA),
				Name:            aws.String(fmt.Sprintf("%s.%s", k, dopts.Fqdn)),
				ResourceRecords: rr,
				TTL:             aws.Int64(int64(route53TTL)),
			}

			if _, err := b.setRecord(rrs, dopts, typeA, t.ID, pID, true); err != nil {
				return err
			}
		}
	}
	return nil
}

// Used to set record to database
func (b *Backend) setRecordToDatabase(rrs *route53.ResourceRecordSet, rType string, tID, pID int64, sub bool) (int64, error) {
	content := make([]string, 0)
	for _, rr := range rrs.ResourceRecords {
		content = append(content, aws.StringValue(rr.Value))
	}

	if rType == typeA && !sub {
		dr := &model.RecordA{
			Type:      1,
			Fqdn:      aws.StringValue(rrs.Name),
			Content:   strings.Join(content, ","),
			TID:       tID,
			CreatedOn: time.Now().Unix(),
		}

		result, _ := database.GetDatabase().QueryA(aws.StringValue(rrs.Name))
		if result != nil && result.Fqdn != "" {
			return database.GetDatabase().UpdateA(dr)
		}
		return database.GetDatabase().InsertA(dr)
	}

	if rType == typeA && sub {
		dr := &model.SubRecordA{
			Type:      2,
			Fqdn:      aws.StringValue(rrs.Name),
			Content:   strings.Join(content, ","),
			PID:       pID,
			CreatedOn: time.Now().Unix(),
		}

		result, _ := database.GetDatabase().QuerySubA(aws.StringValue(rrs.Name))
		if result != nil && result.Fqdn != "" {
			return database.GetDatabase().UpdateSubA(dr)
		}
		return database.GetDatabase().InsertSubA(dr)
	}

	if rType == typeTXT {
		dr := &model.RecordTXT{
			Type:      0,
			Fqdn:      aws.StringValue(rrs.Name),
			Content:   strings.Join(content, ","),
			TID:       tID,
			CreatedOn: time.Now().Unix(),
		}

		result, _ := database.GetDatabase().QueryTXT(aws.StringValue(rrs.Name))
		if result != nil && result.Fqdn != "" {
			return database.GetDatabase().UpdateTXT(dr)
		}
		return database.GetDatabase().InsertTXT(dr)
	}

	return 0, nil
}

// Used to delete record from database
func (b *Backend) deleteRecordFromDatabase(rrs *route53.ResourceRecordSet, rType string, sub bool) error {
	name := strings.TrimRight(aws.StringValue(rrs.Name), ".")
	if rType == typeA && !sub {
		return database.GetDatabase().DeleteA(name)
	}

	if rType == typeA && sub {
		return database.GetDatabase().DeleteSubA(name)
	}

	if rType == typeTXT {
		return database.GetDatabase().DeleteTXT(name)
	}

	return nil
}

// Used to get records
func (b *Backend) getRecords(opts *model.DomainOptions, rType string) (*route53.ListResourceRecordSetsOutput, error) {
	input := route53.ListResourceRecordSetsInput{
		HostedZoneId:    aws.String(b.ZoneID),
		StartRecordName: aws.String(opts.Fqdn),
		StartRecordType: aws.String(rType),
	}

	output, err := b.Svc.ListResourceRecordSets(&input)
	if err != nil {
		return output, errors.Wrapf(err, errEmptyRecord, rType, opts.String())
	}

	return output, nil
}

// Used to set record:
//   parameters:
//     rType: record's type(0: TXT, 1: A, 2: SUB)
//     tID: reference token ID
//     pID: reference parent ID
//     sub: whether is sub domain or not
func (b *Backend) setRecord(rrs *route53.ResourceRecordSet, opts *model.DomainOptions, rType string, tID, pID int64, sub bool) (int64, error) {
	input := route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(b.ZoneID),
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action:            aws.String("UPSERT"),
					ResourceRecordSet: rrs,
				},
			},
		},
	}

	if _, err := b.Svc.ChangeResourceRecordSets(&input); err != nil {
		return 0, errors.Wrapf(err, errUpsertRecord, rType, opts.String())
	}

	// set record to database
	id, err := b.setRecordToDatabase(rrs, rType, tID, pID, sub)
	if err != nil {
		return 0, errors.Wrapf(err, errOperateDatabase, rType, opts.String())
	}

	return id, nil
}

// Used to delete record
//   parameters:
//     rType: record's type(0: TXT, 1: A, 2: SUB)
//     sub: whether is sub domain or not
func (b *Backend) deleteRecord(rrs *route53.ResourceRecordSet, opts *model.DomainOptions, rType string, sub bool) error {
	input := route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(b.ZoneID),
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("DELETE"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name:            rrs.Name,
						Type:            aws.String(rType),
						ResourceRecords: rrs.ResourceRecords,
						TTL:             aws.Int64(int64(route53TTL)),
					},
				},
			},
		},
	}
	if _, err := b.Svc.ChangeResourceRecordSets(&input); err != nil {
		return errors.Wrapf(err, errDeleteRecord, rType, opts.String())
	}

	// delete record from database
	if err := b.deleteRecordFromDatabase(rrs, rType, sub); err != nil {
		return errors.Wrapf(err, errOperateDatabase, rType, opts.String())
	}

	return nil
}

// Used to filter (A,TXT) Records:
//   TXT records:
//     valid:
//       1. Only TXT record which equal to the opts.Fqdn is valid
//   A records:
//     valid:
//       1. A record which equal to the opts.Fqdn is valid
//       2. sub-domain A record which parent is opts.Fqdn is valid
//     not valid:
//       1. wildcard record is not valid
func (b *Backend) filterRecords(rrs []*route53.ResourceRecordSet, opts *model.DomainOptions, rType string) (v bool, a, s, t []*route53.ResourceRecordSet) {
	v = false
	a = make([]*route53.ResourceRecordSet, 0)
	s = make([]*route53.ResourceRecordSet, 0)
	t = make([]*route53.ResourceRecordSet, 0)

	switch rType {
	case typeA:
		for _, rs := range rrs {
			name := strings.TrimRight(aws.StringValue(rs.Name), ".")
			nss := strings.Split(name, ".")
			oss := strings.Split(opts.Fqdn, ".")
			if strings.Contains(name, "*") || strings.Contains(name, "\\052") {
				continue
			}
			if name == opts.Fqdn && aws.StringValue(rs.Type) == rType {
				v = true
				a = append(a, rs)
				continue
			}
			if (len(nss)-len(oss)) == 1 && strings.Contains(name, opts.Fqdn) && aws.StringValue(rs.Type) == rType {
				s = append(s, rs)
				continue
			}
		}
		return
	case typeTXT:
		for _, rs := range rrs {
			name := strings.TrimRight(aws.StringValue(rs.Name), ".")
			if name == strings.TrimRight(opts.Fqdn, ".") && aws.StringValue(rs.Type) == rType {
				v = true
				t = append(t, rs)
				continue
			}
		}
		return
	default:
		return
	}
}

// Used to convert route53 A & sub domain A records to map
func (b *Backend) convertARecords(a, s []*route53.ResourceRecordSet) (aOutput, sOutput map[string][]string) {
	aOutput = make(map[string][]string, 0)
	sOutput = make(map[string][]string, 0)

	for _, rs := range a {
		name := strings.TrimRight(aws.StringValue(rs.Name), ".")
		temp := make([]string, 0)
		for _, r := range rs.ResourceRecords {
			temp = append(temp, aws.StringValue(r.Value))
		}
		aOutput[name] = temp
	}

	for _, rs := range s {
		prefix := strings.Split(aws.StringValue(rs.Name), ".")[0]
		temp := make([]string, 0)
		for _, r := range rs.ResourceRecords {
			temp = append(temp, aws.StringValue(r.Value))
		}
		sOutput[prefix] = temp
	}

	return
}

// Used to find slug name:
//   e.g. yyyy.xxxx.qrn7oq.lb.rancher.cloud => qrn7oq.lb.rancher.cloud
func (b *Backend) findSlugWithZone(fqdn string) string {
	n := len(strings.Split(fqdn, ".")) - (len(strings.Split(b.Zone, ".")))
	ss := strings.SplitAfterN(fqdn, ".", n)
	if len(ss) <= 1 {
		return fqdn
	}
	return ss[1]
}

// Used to generate a random slug
func generateSlug() string {
	return util.RandStringWithSmall(slugLength)
}

// Used to generate a random token
func generateToken() string {
	return util.RandStringWithAll(tokenLength)
}

// Used to convert expiration
func convertExpiration(create time.Time, ttl int) *time.Time {
	duration, _ := time.ParseDuration(fmt.Sprintf("%dns", ttl))
	e := create.Add(duration)
	return &e
}
