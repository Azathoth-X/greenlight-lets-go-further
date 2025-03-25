package data

import (
	"context"
	"database/sql"
	"slices"
	"time"
)

type PermissionsList []string

func (p PermissionsList) Includes(permission string) bool {
	return slices.Contains(p, permission)
}

type PermissionModel struct {
	DB *sql.DB
}

func (m *PermissionModel) GetAllForUser(userid int) (PermissionsList, error) {
	stmt := `SELECT permissions.code 
			 FROM permissions
			 INNER JOIN user_permissions WHERE user_permissions.permissions_id=permissions.id
			 INNER JOIN users WHERE user_permissions.user_id=users.id
			 WHERE users.id=$1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res, err := m.DB.QueryContext(ctx, stmt, userid)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var permissions PermissionsList

	for res.Next() {
		var perm string

		err := res.Scan(&perm)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, perm)
	}

	if err = res.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}
