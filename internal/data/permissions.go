package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Permissions slice, which we will use to hold the permission codes
// like ["movies:read", "movies:write"]
type Permissions []string

// Include is a helper method to check whether the Permissions slice contains
// a specific permissions code
func (p Permissions) Include(code string) bool {
	fmt.Println(p)
	for i := range p {
		fmt.Println("I am here")
		if code == p[i] {
			return true
		}
	}
	return false
}

// PermissionsModel allows us to tap into it via the *sql.DB
type PermissionsModel struct {
	DB *sql.DB
}

// GetAllForUser method allows us to check what permissions a specific user has
// we perform this check by the User ID
func (m PermissionsModel) GetAllForUser(userID int64) (Permissions, error) {

	// define query
	query := `
		SELECT permissions.code
		FROM permissions
		INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
		INNER JOIN users ON users_permissions.user_id = users.id
		WHERE users.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var permissions Permissions

	for rows.Next() {

		var permission string

		err := rows.Scan(&permission)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil

}
