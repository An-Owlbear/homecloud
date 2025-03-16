package auth

import (
	"context"

	kratos "github.com/ory/kratos-client-go"
)

// ListUsers lists all users in the system. This isn't paginated is it isn't expected a server like this would
func ListUsers(ctx context.Context, kratosAdmin kratos.IdentityAPI) (users []kratos.Identity, err error) {
	users, _, err = kratosAdmin.ListIdentities(ctx).Execute()
	if err != nil {
		return
	}
	return
}

// DeleteUser deletes the specified user
func DeleteUser(ctx context.Context, kratosAdmin kratos.IdentityAPI, userId string) error {
	_, err := kratosAdmin.DeleteIdentity(ctx, userId).Execute()
	return err
}
