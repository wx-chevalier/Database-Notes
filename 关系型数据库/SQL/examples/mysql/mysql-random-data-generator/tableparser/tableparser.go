package tableparser

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}

// Scan implements the Scanner interface.
func (nt *NullTime) Scan(value interface{}) error {
	nt.Time, nt.Valid = value.(time.Time)
	return nil
}

// Value implements the driver Valuer interface.
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

// Table holds the table definition with all fields, indexes and triggers
type Table struct {
	Schema  string
	Name    string
	Fields  []Field
	Indexes map[string]Index
	//TODO Include complete indexes information
	Constraints []Constraint
	Triggers    []Trigger
	//
	conn *sql.DB
}

// Index holds the basic index information
type Index struct {
	Name    string
	Fields  []string
	Unique  bool
	Visible bool
}

// IndexField holds raw index information as defined in INFORMATION_SCHEMA table
type IndexField struct {
	KeyName      string
	SeqInIndex   int
	ColumnName   string
	Collation    sql.NullString
	Cardinality  sql.NullInt64
	SubPart      sql.NullInt64
	Packed       sql.NullString
	Null         string
	IndexType    string
	Comment      string
	IndexComment string
	NonUnique    bool
	Visible      bool // MySQL 8.0+
}

// Constraint holds Foreign Keys information
type Constraint struct {
	ConstraintName        string
	ColumnName            string
	ReferencedTableSchema string
	ReferencedTableName   string
	ReferencedColumnName  string
}

// Field holds raw field information as defined in INFORMATION_SCHEMA
type Field struct {
	TableCatalog           string
	TableSchema            string
	TableName              string
	ColumnName             string
	OrdinalPosition        int
	ColumnDefault          sql.NullString
	IsNullable             bool
	DataType               string
	CharacterMaximumLength sql.NullInt64
	CharacterOctetLength   sql.NullInt64
	NumericPrecision       sql.NullInt64
	NumericScale           sql.NullInt64
	DatetimePrecision      sql.NullInt64
	CharacterSetName       sql.NullString
	CollationName          sql.NullString
	ColumnType             string
	ColumnKey              string
	Extra                  string
	Privileges             string
	ColumnComment          string
	GenerationExpression   string
	SetEnumVals            []string
	Constraint             *Constraint
	SrsID                  sql.NullString
}

// Trigger holds raw trigger information as defined in INFORMATION_SCHEMA
type Trigger struct {
	Trigger             string
	Event               string
	Table               string
	Statement           string
	Timing              string
	Created             NullTime
	SQLMode             string
	Definer             string
	CharacterSetClient  string
	CollationConnection string
	DatabaseCollation   string
}

func NewTable(db *sql.DB, schema, tableName string) (*Table, error) {
	table := &Table{
		Schema: url.QueryEscape(schema),
		Name:   url.QueryEscape(tableName),
		conn:   db,
	}

	var err error
	table.Indexes, err = getIndexes(db, table.Schema, table.Name)
	if err != nil {
		return nil, err
	}
	table.Constraints, err = getConstraints(db, table.Schema, table.Name)
	if err != nil {
		return nil, err
	}
	table.Triggers, err = getTriggers(db, table.Schema, table.Name)
	if err != nil {
		return nil, err
	}

	err = table.parse()
	if err != nil {
		return nil, err
	}
	table.conn = nil // to save memory since it is not going to be used again
	return table, nil
}

func (t *Table) parse() error {
	//                           +--------------------------- field type
	//                           |          +---------------- field size / enum values: decimal(10,2) or enum('a','b')
	//                           |          |     +---------- extra info (unsigned, etc)
	//                           |          |     |
	re := regexp.MustCompile(`^(.*?)(?:\((.*?)\)(.*))?$`)

	query := "SELECT * FROM `information_schema`.`COLUMNS` WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? ORDER BY ORDINAL_POSITION"

	constraints := constraintsAsMap(t.Constraints)

	rows, err := t.conn.Query(query, t.Schema, t.Name)
	if err != nil {
		return err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return errors.Wrap(err, "Cannot get column names")
	}

	for rows.Next() {
		var f Field
		var allowNull string
		fields := makeScanRecipients(&f, &allowNull, cols)
		err := rows.Scan(fields...)
		if err != nil {
			log.Errorf("Cannot get table fields: %s", err)
			continue
		}

		allowedValues := []string{}
		if f.DataType == "enum" || f.DataType == "set" {
			m := re.FindStringSubmatch(f.ColumnType)
			if len(m) < 2 {
				continue
			}
			vals := strings.Split(m[2], ",")
			for _, val := range vals {
				val = strings.TrimPrefix(val, "'")
				val = strings.TrimSuffix(val, "'")
				allowedValues = append(allowedValues, val)
			}
		}

		f.SetEnumVals = allowedValues
		f.IsNullable = allowNull == "YES"
		f.Constraint = constraints[f.ColumnName]

		t.Fields = append(t.Fields, f)
	}

	if rows.Err() != nil {
		return rows.Err()
	}
	return nil
}

func makeScanRecipients(f *Field, allowNull *string, cols []string) []interface{} {
        fields := []interface{}{
                &f.TableCatalog,
                &f.TableSchema,
                &f.TableName,
                &f.ColumnName,
                &f.OrdinalPosition,
                &f.ColumnDefault,
                &allowNull,
                &f.DataType,
                &f.CharacterMaximumLength,
                &f.CharacterOctetLength,
                &f.NumericPrecision,
                &f.NumericScale,
        }

        if len(cols) > 19 { // MySQL 5.5 does not have "DATETIME_PRECISION" field
        fields = append(fields, &f.DatetimePrecision)
        }

        fields = append(fields, &f.CharacterSetName, &f.CollationName, &f.ColumnType, &f.ColumnKey, &f.Extra, &f.Privileges, &f.ColumnComment)

        if len(cols) > 20 && cols[20] == "GENERATION_EXPRESSION" { // MySQL 5.7+ "GENERATION_EXPRESSION" field
                fields = append(fields, &f.GenerationExpression)
        }
        if len(cols) > 21 && cols[21] == "SRS_ID" { // MySQL 8.0+ "SRS ID" field
                fields = append(fields, &f.SrsID)
        }

        return fields
}

// FieldNames returns an string array with the table's field names
func (t *Table) FieldNames() []string {
	fields := []string{}
	for _, field := range t.Fields {
		fields = append(fields, field.ColumnName)
	}
	return fields
}

func getIndexes(db *sql.DB, schema, tableName string) (map[string]Index, error) {
	query := fmt.Sprintf("SHOW INDEXES FROM `%s`.`%s`", schema, tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	indexes := make(map[string]Index)

	for rows.Next() {
		var i IndexField
		var table, visible string
		fields := []interface{}{&table, &i.NonUnique, &i.KeyName, &i.SeqInIndex,
			&i.ColumnName, &i.Collation, &i.Cardinality, &i.SubPart,
			&i.Packed, &i.Null, &i.IndexType, &i.Comment, &i.IndexComment,
		}

		cols, err := rows.Columns()
		if err == nil && len(cols) == 14 && cols[13] == "Visible" {
			fields = append(fields, &visible)
		}

		err = rows.Scan(fields...)
		if err != nil {
			return nil, fmt.Errorf("cannot read indexes: %s", err)
		}
		if index, ok := indexes[i.KeyName]; !ok {
			indexes[i.KeyName] = Index{
				Name:    i.KeyName,
				Unique:  !i.NonUnique,
				Fields:  []string{i.ColumnName},
				Visible: visible == "YES" || visible == "",
			}

		} else {
			index.Fields = append(index.Fields, i.ColumnName)
			index.Unique = index.Unique || !i.NonUnique
		}
	}
	if err := rows.Close(); err != nil {
		return nil, errors.Wrap(err, "Cannot close query rows at getIndexes")
	}

	return indexes, nil
}

func getConstraints(db *sql.DB, schema, tableName string) ([]Constraint, error) {
	query := "SELECT tc.CONSTRAINT_NAME, " +
		"kcu.COLUMN_NAME, " +
		"kcu.REFERENCED_TABLE_SCHEMA, " +
		"kcu.REFERENCED_TABLE_NAME, " +
		"kcu.REFERENCED_COLUMN_NAME " +
		"FROM information_schema.TABLE_CONSTRAINTS tc " +
		"LEFT JOIN information_schema.KEY_COLUMN_USAGE kcu " +
		"ON tc.CONSTRAINT_NAME = kcu.CONSTRAINT_NAME " +
		"WHERE tc.CONSTRAINT_TYPE = 'FOREIGN KEY' " +
		fmt.Sprintf("AND tc.TABLE_SCHEMA = '%s' ", schema) +
		fmt.Sprintf("AND tc.TABLE_NAME = '%s'", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	constraints := []Constraint{}

	for rows.Next() {
		var c Constraint
		err := rows.Scan(&c.ConstraintName, &c.ColumnName, &c.ReferencedTableSchema,
			&c.ReferencedTableName, &c.ReferencedColumnName)
		if err != nil {
			return nil, fmt.Errorf("cannot read constraints: %s", err)
		}
		constraints = append(constraints, c)
	}

	return constraints, nil
}

func getTriggers(db *sql.DB, schema, tableName string) ([]Trigger, error) {
	query := fmt.Sprintf("SHOW TRIGGERS FROM `%s` LIKE '%s'", schema, tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	triggers := []Trigger{}

	for rows.Next() {
		var t Trigger
		err := rows.Scan(&t.Trigger, &t.Event, &t.Table, &t.Statement, &t.Timing,
			&t.Created, &t.SQLMode, &t.Definer, &t.CharacterSetClient, &t.CollationConnection,
			&t.DatabaseCollation)
		if err != nil {
			return nil, fmt.Errorf("cannot read trigger: %s", err)
		}
		triggers = append(triggers, t)
	}

	return triggers, nil
}

func constraintsAsMap(constraints []Constraint) map[string]*Constraint {
	m := make(map[string]*Constraint)
	for _, c := range constraints {
		m[c.ColumnName] = &Constraint{
			ConstraintName:        c.ConstraintName,
			ColumnName:            c.ColumnName,
			ReferencedTableSchema: c.ReferencedTableSchema,
			ReferencedTableName:   c.ReferencedTableName,
			ReferencedColumnName:  c.ReferencedColumnName,
		}
	}
	return m
}
