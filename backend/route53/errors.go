package route53

const (
	errDeleteAFromDatabase       = "failed to delete A record %s from database"
	errDeleteRecordsFromDatabase = "failed to delete %s record %s from database"
	errDeleteRoute53Record       = "failed to delete route53 %s record: %s"
	errExistRecord               = "%s record: %s already exist"
	errFilterRecords             = "failed to filter %s records: %s"
	errGenerateName              = "failed to generate valid record: %s"
	errInsertFrozenToDatabase    = "failed to insert %s's frozen to database"
	errInsertRecordToDatabase    = "failed to insert %s record: %s to database"
	errInsertTokenToDatabase     = "failed to insert %s's token to database"
	errNoRoute53Record           = "failed to found route53 %s record: %s"
	errNotValidGenerateName      = "generate name %s is already exist, will try another"
	errParseFlag                 = "failed to parse flag: %s"
	errQueryAFromDatabase        = "failed to query %s's A record from database"
	errQueryTokenFromDatabase    = "failed to query %s's token record from database"
	errQueryTXTFromDatabase      = "failed to query %s's TXT record from database"
	errQueryCNAMEFromDatabase    = "failed to query %s's CNAME record from database"
	errRenewFrozenFromDatabase   = "failed to renew %s's frozen record from database"
	errRenewTokenFromDatabase    = "failed to renew %s's token record from database"
	errUpsertRoute53Record       = "failed to upsert route53 %s record: %s"
)
