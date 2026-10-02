package database

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// UUIDToPg converts a google/uuid.UUID into pgtype.UUID
func UUIDToPg(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: id,
		Valid: id != uuid.Nil,
	}
}

// PgToUUID converts a pgtype.UUID into google/uuid.UUID
func PgToUUID(pgID pgtype.UUID) uuid.UUID {
	if !pgID.Valid {
		return uuid.Nil
	}
	return pgID.Bytes
}

// Float64ToNumeric converts a float64 into pgtype.Numeric
func Float64ToNumeric(f float64) pgtype.Numeric {
	var num pgtype.Numeric
	_ = num.Scan(strconv.FormatFloat(f, 'f', -1, 64))
	return num
}

// Float64PtrToNumeric converts a *float64 into pgtype.Numeric
func Float64PtrToNumeric(f *float64) pgtype.Numeric {
	if f == nil {
		return pgtype.Numeric{Valid: false}
	}
	return Float64ToNumeric(*f)
}

// NumericToFloat64Ptr converts pgtype.Numeric to *float64
func NumericToFloat64Ptr(n pgtype.Numeric) *float64 {
	if !n.Valid {
		return nil
	}
	val, err := n.Float64Value()
	if err != nil || !val.Valid {
		if n.Int != nil {
			f := float64(n.Int.Int64())
			if n.Exp < 0 {
				for i := 0; i < int(-n.Exp); i++ {
					f /= 10.0
				}
			} else if n.Exp > 0 {
				for i := 0; i < int(n.Exp); i++ {
					f *= 10.0
				}
			}
			return &f
		}
		return nil
	}
	return &val.Float64
}

// StringToPgText converts string to pgtype.Text
func StringToPgText(s string) pgtype.Text {
	return pgtype.Text{
		String: s,
		Valid:  s != "",
	}
}
