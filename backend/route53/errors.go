package route53

const (
	errParseFlag            = "failed to parse flag: %s"
	errEmptyRecord          = "failed to found %s record: %s"
	errExistRecord          = "%s record: %s already exist"
	errUpsertRecord         = "failed to upsert %s record: %s"
	errDeleteRecord         = "failed to delete %s record: %s"
	errOperateDatabase      = "failed to operate database %s record: %s"
	errGenerateName         = "failed to generate valid record: %s"
	errNotValidGenerateName = "generate name %s is already exist, will try another"
)
