package internalversion

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	userv1 "github.com/openshift/api/user/v1"
	"github.com/openshift/oauth-apiserver/pkg/printers"
	userapi "github.com/openshift/oauth-apiserver/pkg/user/apis/user"
)

func AddUserOpenShiftHandler(h printers.PrintHandler) {
	addUser(h)
	addIdentity(h)
	addUserIdentityMapping(h)
	addGroup(h)
}

func addUser(h printers.PrintHandler) {
	userColumnsDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "UID", Type: "string", Description: metav1.ObjectMeta{}.SwaggerDoc()["uid"]},
		{Name: "Full Name", Type: "string", Description: userv1.User{}.SwaggerDoc()["fullName"]},
		{Name: "Identities", Type: "string", Description: userv1.User{}.SwaggerDoc()["identities"]},
	}
	if err := h.TableHandler(userColumnsDefinitions, printUser); err != nil {
		panic(err)
	}
	if err := h.TableHandler(userColumnsDefinitions, printUserList); err != nil {
		panic(err)
	}
}

// formatResourceName receives a resource kind, name, and boolean specifying
// whether or not to update the current name to "kind/name"
func formatResourceName(kind schema.GroupKind, name string, withKind bool) string {
	if !withKind || kind.Empty() {
		return name
	}

	return strings.ToLower(kind.String()) + "/" + name
}

func printUser(user *userapi.User, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: user},
	}

	row.Cells = append(row.Cells,
		user.Name,
		user.UID,
		user.FullName,
		strings.Join(user.Identities, ", "),
	)

	return []metav1.TableRow{row}, nil
}

func printUserList(userList *userapi.UserList, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(userList.Items))
	for i := range userList.Items {
		r, err := printUser(&userList.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func addIdentity(h printers.PrintHandler) {
	identityColumnsDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "IDP Name", Type: "string", Format: "name", Description: userv1.Identity{}.SwaggerDoc()["providerName"]},
		{Name: "IDP User Name", Type: "string", Format: "name", Description: userv1.Identity{}.SwaggerDoc()["providerUserName"]},
		{Name: "User Name", Type: "string", Format: "name", Description: userv1.Identity{}.SwaggerDoc()["user"]},
		{Name: "User UID", Type: "string", Description: metav1.ObjectMeta{}.SwaggerDoc()["uid"]},
	}
	if err := h.TableHandler(identityColumnsDefinitions, printIdentity); err != nil {
		panic(err)
	}
	if err := h.TableHandler(identityColumnsDefinitions, printIdentityList); err != nil {
		panic(err)
	}
}

func printIdentity(identity *userapi.Identity, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: identity},
	}

	row.Cells = append(row.Cells,
		identity.Name,
		identity.ProviderName,
		identity.ProviderUserName,
		identity.User.Name,
		identity.User.UID,
	)

	return []metav1.TableRow{row}, nil
}

func printIdentityList(identityList *userapi.IdentityList, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(identityList.Items))
	for i := range identityList.Items {
		r, err := printIdentity(&identityList.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func addUserIdentityMapping(h printers.PrintHandler) {
	identityColumnsDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Identity", Type: "string", Description: userv1.UserIdentityMapping{}.SwaggerDoc()["identity"]},
		{Name: "User Name", Type: "string", Description: userv1.UserIdentityMapping{}.SwaggerDoc()["user"]},
		{Name: "User UID", Type: "string", Description: metav1.ObjectMeta{}.SwaggerDoc()["uid"]},
	}
	if err := h.TableHandler(identityColumnsDefinitions, printUserIdentityMapping); err != nil {
		panic(err)
	}
}

func printUserIdentityMapping(mapping *userapi.UserIdentityMapping, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: mapping},
	}

	row.Cells = append(row.Cells,
		mapping.Name,
		mapping.Identity.Name,
		mapping.User.Name,
		mapping.User.UID,
	)

	return []metav1.TableRow{row}, nil
}

func addGroup(h printers.PrintHandler) {
	groupColumnsDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Users", Type: "string", Description: userv1.Group{}.SwaggerDoc()["users"]},
	}

	if err := h.TableHandler(groupColumnsDefinitions, printGroup); err != nil {
		panic(err)
	}
	if err := h.TableHandler(groupColumnsDefinitions, printGroupList); err != nil {
		panic(err)
	}
}

func printGroup(group *userapi.Group, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: group},
	}

	row.Cells = append(row.Cells,
		group.Name,
		strings.Join(group.Users, ", "),
	)

	return []metav1.TableRow{row}, nil
}

func printGroupList(groupList *userapi.GroupList, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(groupList.Items))
	for i := range groupList.Items {
		r, err := printGroup(&groupList.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}
