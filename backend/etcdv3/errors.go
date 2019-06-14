package etcdv3

const (
	errEmptyRecord            = "failed to found %s record: %s"
	errMultiRecords           = "multiple %s records: %s"
	errLookupRecords          = "failed to lookup %s record: %s"
	errExistSlug              = "slug name %s can not be used, try another"
	errGrantLease             = "failed to grant lease"
	errKeepaliveOnce          = "failed to keepaliveOnce with lease %d"
	errSetRecordWithLease     = "failed to set %s record %s with lease"
	errSetSubRecordsWithLease = "failed to set sub %s records %s with lease"
	errSyncRecords            = "failed to sync %s records: %s"
	errSyncSubRecords         = "failed to sync sub %s records: %s"
	errDeleteRecord           = "failed to delete %s record: %s"
	errNotValidFqdn           = "not valid fqdn: %s"
)
