package api

// SQLReviewRuleType is the type of schema rule.
type SQLReviewRuleType string

// SQLReviewRuleLevel is the error level for SQL review rule.
type SQLReviewRuleLevel string

const (
	// SchemaRuleTableNaming enforce the table name format.
	SchemaRuleTableNaming SQLReviewRuleType = "naming.table"
	// SchemaRuleColumnNaming enforce the column name format.
	SchemaRuleColumnNaming SQLReviewRuleType = "naming.column"
	// SchemaRulePKNaming enforce the primary key name format.
	SchemaRulePKNaming SQLReviewRuleType = "naming.index.pk"
	// SchemaRuleUKNaming enforce the unique key name format.
	SchemaRuleUKNaming SQLReviewRuleType = "naming.index.uk"
	// SchemaRuleFKNaming enforce the foreign key name format.
	SchemaRuleFKNaming SQLReviewRuleType = "naming.index.fk"
	// SchemaRuleIDXNaming enforce the index name format.
	SchemaRuleIDXNaming SQLReviewRuleType = "naming.index.idx"
	// SchemaRuleAutoIncrementColumnNaming enforce the auto_increment column name format.
	SchemaRuleAutoIncrementColumnNaming SQLReviewRuleType = "naming.column.auto-increment"

	// SchemaRuleStatementNoSelectAll disallow 'SELECT *'.
	SchemaRuleStatementNoSelectAll SQLReviewRuleType = "statement.select.no-select-all"
	// SchemaRuleStatementRequireWhere require 'WHERE' clause.
	SchemaRuleStatementRequireWhere SQLReviewRuleType = "statement.where.require"
	// SchemaRuleStatementNoLeadingWildcardLike disallow leading '%' in LIKE, e.g. LIKE foo = '%x' is not allowed.
	SchemaRuleStatementNoLeadingWildcardLike SQLReviewRuleType = "statement.where.no-leading-wildcard-like"
	// SchemaRuleStatementDisallowCommit disallow using commit in the issue.
	SchemaRuleStatementDisallowCommit SQLReviewRuleType = "statement.disallow-commit"
	// SchemaRuleStatementDisallowLimit disallow the LIMIT clause in INSERT, DELETE and UPDATE statements.
	SchemaRuleStatementDisallowLimit SQLReviewRuleType = "statement.disallow-limit"
	// SchemaRuleStatementDisallowOrderBy disallow the ORDER BY clause in DELETE and UPDATE statements.
	SchemaRuleStatementDisallowOrderBy SQLReviewRuleType = "statement.disallow-order-by"
	// SchemaRuleStatementMergeAlterTable disallow redundant ALTER TABLE statements.
	SchemaRuleStatementMergeAlterTable SQLReviewRuleType = "statement.merge-alter-table"
	// SchemaRuleStatementInsertRowLimit enforce the insert row limit.
	SchemaRuleStatementInsertRowLimit SQLReviewRuleType = "statement.insert.row-limit"
	// SchemaRuleStatementInsertMustSpecifyColumn enforce the insert column specified.
	SchemaRuleStatementInsertMustSpecifyColumn SQLReviewRuleType = "statement.insert.must-specify-column"
	// SchemaRuleStatementInsertDisallowOrderByRand disallow the order by rand in the INSERT statement.
	SchemaRuleStatementInsertDisallowOrderByRand SQLReviewRuleType = "statement.insert.disallow-order-by-rand"
	// SchemaRuleStatementAffectedRowLimit enforce the UPDATE/DELETE affected row limit.
	SchemaRuleStatementAffectedRowLimit SQLReviewRuleType = "statement.affected-row-limit"
	// SchemaRuleStatementDMLDryRun dry run the dml.
	SchemaRuleStatementDMLDryRun SQLReviewRuleType = "statement.dml-dry-run"

	// SchemaRuleTableRequirePK require the table to have a primary key.
	SchemaRuleTableRequirePK SQLReviewRuleType = "table.require-pk"
	// SchemaRuleTableNoFK require the table disallow the foreign key.
	SchemaRuleTableNoFK SQLReviewRuleType = "table.no-foreign-key"
	// SchemaRuleTableDropNamingConvention require only the table following the naming convention can be deleted.
	SchemaRuleTableDropNamingConvention SQLReviewRuleType = "table.drop-naming-convention"
	// SchemaRuleTableCommentConvention enforce the table comment convention.
	SchemaRuleTableCommentConvention SQLReviewRuleType = "table.comment"
	// SchemaRuleTableDisallowPartition disallow the table partition.
	SchemaRuleTableDisallowPartition SQLReviewRuleType = "table.disallow-partition"

	// SchemaRuleRequiredColumn enforce the required columns in each table.
	SchemaRuleRequiredColumn SQLReviewRuleType = "column.required"
	// SchemaRuleColumnNotNull enforce the columns cannot have NULL value.
	SchemaRuleColumnNotNull SQLReviewRuleType = "column.no-null"
	// SchemaRuleColumnDisallowChangeType disallow change column type.
	SchemaRuleColumnDisallowChangeType SQLReviewRuleType = "column.disallow-change-type"
	// SchemaRuleColumnSetDefaultForNotNull require the not null column to set default value.
	SchemaRuleColumnSetDefaultForNotNull SQLReviewRuleType = "column.set-default-for-not-null"
	// SchemaRuleColumnDisallowChange disallow CHANGE COLUMN statement.
	SchemaRuleColumnDisallowChange SQLReviewRuleType = "column.disallow-change"
	// SchemaRuleColumnDisallowChangingOrder disallow changing column order.
	SchemaRuleColumnDisallowChangingOrder SQLReviewRuleType = "column.disallow-changing-order"
	// SchemaRuleColumnCommentConvention enforce the column comment convention.
	SchemaRuleColumnCommentConvention SQLReviewRuleType = "column.comment"
	// SchemaRuleColumnAutoIncrementMustInteger require the auto-increment column to be integer.
	SchemaRuleColumnAutoIncrementMustInteger SQLReviewRuleType = "column.auto-increment-must-integer"
	// SchemaRuleColumnTypeDisallowList enforce the column type disallow list.
	SchemaRuleColumnTypeDisallowList SQLReviewRuleType = "column.type-disallow-list"
	// SchemaRuleColumnDisallowSetCharset disallow set column charset.
	SchemaRuleColumnDisallowSetCharset SQLReviewRuleType = "column.disallow-set-charset"
	// SchemaRuleColumnMaximumCharacterLength enforce the maximum character length.
	SchemaRuleColumnMaximumCharacterLength SQLReviewRuleType = "column.maximum-character-length"
	// SchemaRuleColumnAutoIncrementInitialValue enforce the initial auto-increment value.
	SchemaRuleColumnAutoIncrementInitialValue SQLReviewRuleType = "column.auto-increment-initial-value"
	// SchemaRuleColumnAutoIncrementMustUnsigned enforce the auto-increment column to be unsigned.
	SchemaRuleColumnAutoIncrementMustUnsigned SQLReviewRuleType = "column.auto-increment-must-unsigned"
	// SchemaRuleCurrentTimeColumnCountLimit enforce the current column count limit.
	SchemaRuleCurrentTimeColumnCountLimit SQLReviewRuleType = "column.current-time-count-limit"
	// SchemaRuleColumnRequireDefault enforce the column default.
	SchemaRuleColumnRequireDefault SQLReviewRuleType = "column.require-default"

	// SchemaRuleSchemaBackwardCompatibility enforce the MySQL and TiDB support check whether the schema change is backward compatible.
	SchemaRuleSchemaBackwardCompatibility SQLReviewRuleType = "schema.backward-compatibility"

	// SchemaRuleDropEmptyDatabase enforce the MySQL and TiDB support check if the database is empty before users drop it.
	SchemaRuleDropEmptyDatabase SQLReviewRuleType = "database.drop-empty-database"

	// SchemaRuleIndexNoDuplicateColumn require the index no duplicate column.
	SchemaRuleIndexNoDuplicateColumn SQLReviewRuleType = "index.no-duplicate-column"
	// SchemaRuleIndexKeyNumberLimit enforce the index key number limit.
	SchemaRuleIndexKeyNumberLimit SQLReviewRuleType = "index.key-number-limit"
	// SchemaRuleIndexPKTypeLimit enforce the type restriction of columns in primary key.
	SchemaRuleIndexPKTypeLimit SQLReviewRuleType = "index.pk-type-limit"
	// SchemaRuleIndexTypeNoBlob enforce the type restriction of columns in index.
	SchemaRuleIndexTypeNoBlob SQLReviewRuleType = "index.type-no-blob"
	// SchemaRuleIndexTotalNumberLimit enforce the index total number limit.
	SchemaRuleIndexTotalNumberLimit SQLReviewRuleType = "index.total-number-limit"
	// SchemaRuleIndexPrimaryKeyTypeAllowlist enforce the primary key type allowlist.
	SchemaRuleIndexPrimaryKeyTypeAllowlist SQLReviewRuleType = "index.primary-key-type-allowlist"

	// SchemaRuleCharsetAllowlist enforce the charset allowlist.
	SchemaRuleCharsetAllowlist SQLReviewRuleType = "system.charset.allowlist"

	// SchemaRuleCollationAllowlist enforce the collation allowlist.
	SchemaRuleCollationAllowlist SQLReviewRuleType = "system.collation.allowlist"

	// SchemaRuleCommentLength limit comment length.
	SchemaRuleCommentLength SQLReviewRuleType = "comment.length"

	// SQLReviewRuleLevelError is the error level for SQL review rule.
	SQLReviewRuleLevelError SQLReviewRuleLevel = "ERROR"
	// SQLReviewRuleLevelWarning is the warning level for SQL review rule.
	SQLReviewRuleLevelWarning SQLReviewRuleLevel = "WARNING"
	// SQLReviewRuleLevelDisabled is the disabled level for SQL review rule.
	SQLReviewRuleLevelDisabled SQLReviewRuleLevel = "DISABLED"
)

// NamingRulePayload is the payload for naming rule.
type NamingRulePayload struct {
	MaxLength int    `json:"maxLength"`
	Format    string `json:"format"`
}

// StringArrayTypeRulePayload is the payload for rules with string array value.
type StringArrayTypeRulePayload struct {
	List []string `json:"list"`
}

// RequiredColumnRulePayload is the payload for required column rule.
type RequiredColumnRulePayload struct {
	ColumnList []string `json:"columnList"`
}

// CommentConventionRulePayload is the payload for comment convention rule.
type CommentConventionRulePayload struct {
	Required  bool `json:"required"`
	MaxLength int  `json:"maxLength"`
}

// NumberTypeRulePayload is the number type payload.
type NumberTypeRulePayload struct {
	Number int `json:"number"`
}

// SQLReviewRule is the API message for SQL review rule.
type SQLReviewRule struct {
	Type    SQLReviewRuleType  `json:"type"`
	Level   SQLReviewRuleLevel `json:"level"`
	Payload string             `json:"payload"`
}
