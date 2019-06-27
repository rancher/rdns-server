package etcdv3

const (
	errDeleteRecord           = "failed to delete %s record: %s"
	errEmptyRecord            = "failed to found %s record: %s"
	errExistSlug              = "slug name %s can not be used, try another"
	errGrantLease             = "failed to grant lease"
	errSetRecordWithLease     = "failed to set %s record %s with lease %d"
	errSyncRecords            = "failed to sync %s records: %s"
	errSyncSubRecords         = "failed to sync sub %s records: %s"
	errSetSubRecordsWithLease = "failed to set sub %s records %s with lease %d"
	errKeepaliveOnce          = "failed to keepaliveOnce with lease %d"
	errLookupRecords          = "failed to lookup %s record: %s"
	errMultiRecords           = "multiple %s records: %s"
	errNoLookupResults        = "no lookup results for %s record: %s"
	errNotValidDomainName     = "not valid domain name: %s"
)
