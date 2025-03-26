package data

import (
	"context"
	"database/sql"
	"slices"
	"time"

	"github.com/lib/pq"
)

type PermissionsList []string

func (p PermissionsList) Includes(permission string) bool {
	return slices.Contains(p, permission)
}

type PermissionModel struct {
	DB *sql.DB
}

func (m *PermissionModel) GetAllForUser(userid int64) (PermissionsList, error) {
	stmt := ` SELECT permissions.code
			FROM permissions
			INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
			INNER JOIN users ON users_permissions.user_id = users.id
			WHERE users.id = $1`

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

func (m *PermissionModel) AddForUser(userid int64, codes ...string) error {
	stmt := ` INSERT INTO users_permissions
			SELECT $1, permissions.id FROM permissions WHERE permissions.code = ANY($2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, userid, pq.Array(codes))
	return err
}
