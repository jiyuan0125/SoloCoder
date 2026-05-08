package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/example/rbac-permission-tree/internal/api"
)

type Client struct {
	BaseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{BaseURL: strings.TrimSuffix(baseURL, "/")}
}

func (c *Client) do(method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		var errResp api.ErrorResponse
		if json.Unmarshal(data, &errResp) == nil {
			return nil, fmt.Errorf("%s", errResp.Error)
		}
		return nil, fmt.Errorf("http error: %d", resp.StatusCode)
	}

	return data, nil
}

func (c *Client) RoleList() error {
	data, err := c.do(http.MethodGet, "/roles", nil)
	if err != nil {
		return err
	}

	var resp api.ListRolesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	for _, role := range resp.Roles {
		fmt.Printf("ID: %s\n", role.ID)
		fmt.Printf("  Name: %s\n", role.Name)
		fmt.Printf("  Description: %s\n", role.Description)
		fmt.Println()
	}
	return nil
}

func (c *Client) RoleCreate(id, name, desc string) error {
	req := api.CreateRoleRequest{
		ID:          id,
		Name:        name,
		Description: desc,
	}
	_, err := c.do(http.MethodPost, "/roles", req)
	if err != nil {
		return err
	}
	fmt.Printf("Role %s created\n", id)
	return nil
}

func (c *Client) RoleGet(id string) error {
	data, err := c.do(http.MethodGet, "/roles/"+id, nil)
	if err != nil {
		return err
	}

	var role api.Role
	if err := json.Unmarshal(data, &role); err != nil {
		return err
	}

	fmt.Printf("ID: %s\n", role.ID)
	fmt.Printf("Name: %s\n", role.Name)
	fmt.Printf("Description: %s\n", role.Description)
	return nil
}

func (c *Client) RoleUpdate(id, name, desc string) error {
	req := api.UpdateRoleRequest{
		Name:        name,
		Description: desc,
	}
	_, err := c.do(http.MethodPut, "/roles/"+id, req)
	if err != nil {
		return err
	}
	fmt.Printf("Role %s updated\n", id)
	return nil
}

func (c *Client) RoleDelete(id string) error {
	_, err := c.do(http.MethodDelete, "/roles/"+id, nil)
	if err != nil {
		return err
	}
	fmt.Printf("Role %s deleted\n", id)
	return nil
}

func (c *Client) PermList(roleID string) error {
	data, err := c.do(http.MethodGet, "/roles/"+roleID+"/permissions", nil)
	if err != nil {
		return err
	}

	var resp api.GetRolePermissionsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	for _, perm := range resp.Permissions {
		fmt.Printf("  %s:%s (%s)\n", perm.Resource, perm.Action, perm.Scope)
	}
	return nil
}

func (c *Client) PermAdd(roleID, resource, action, scope string) error {
	req := api.AddPermissionRequest{
		Permissions: []api.Permission{
			{
				Resource: resource,
				Action:   action,
				Scope:    api.Scope(scope),
			},
		},
	}
	_, err := c.do(http.MethodPost, "/roles/"+roleID+"/permissions", req)
	if err != nil {
		return err
	}
	fmt.Printf("Permission %s:%s added to role %s\n", resource, action, roleID)
	return nil
}

func (c *Client) PermRemove(roleID, resource, action string) error {
	req := api.RemovePermissionRequest{
		Resource: resource,
		Action:   action,
	}
	_, err := c.do(http.MethodDelete, "/roles/"+roleID+"/permissions", req)
	if err != nil {
		return err
	}
	fmt.Printf("Permission %s:%s removed from role %s\n", resource, action, roleID)
	return nil
}

func (c *Client) ParentAdd(childID, parentID string) error {
	req := api.AddParentRequest{ParentID: parentID}
	_, err := c.do(http.MethodPost, "/roles/"+childID+"/parents", req)
	if err != nil {
		return err
	}
	fmt.Printf("Role %s now inherits from %s\n", childID, parentID)
	return nil
}

func (c *Client) ParentRemove(childID, parentID string) error {
	req := api.RemoveParentRequest{ParentID: parentID}
	_, err := c.do(http.MethodDelete, "/roles/"+childID+"/parents", req)
	if err != nil {
		return err
	}
	fmt.Printf("Removed inheritance %s -> %s\n", childID, parentID)
	return nil
}

func (c *Client) ParentList(roleID string) error {
	data, err := c.do(http.MethodGet, "/roles/"+roleID+"/parents", nil)
	if err != nil {
		return err
	}

	var resp api.GetRoleParentsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	for _, parent := range resp.Parents {
		fmt.Println(parent)
	}
	return nil
}

func (c *Client) UserRoleAdd(userID, roleID string) error {
	req := api.AddRoleToUserRequest{RoleID: roleID}
	_, err := c.do(http.MethodPost, "/users/"+userID+"/roles", req)
	if err != nil {
		return err
	}
	fmt.Printf("Role %s assigned to user %s\n", roleID, userID)
	return nil
}

func (c *Client) UserRoleRemove(userID, roleID string) error {
	req := api.RemoveRoleFromUserRequest{RoleID: roleID}
	_, err := c.do(http.MethodDelete, "/users/"+userID+"/roles", req)
	if err != nil {
		return err
	}
	fmt.Printf("Role %s removed from user %s\n", roleID, userID)
	return nil
}

func (c *Client) UserRoleList(userID string) error {
	data, err := c.do(http.MethodGet, "/users/"+userID+"/roles", nil)
	if err != nil {
		return err
	}

	var resp api.GetUserRolesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	for _, role := range resp.Roles {
		fmt.Println(role)
	}
	return nil
}

func (c *Client) Check(userID, resource, action string) error {
	req := api.CheckPermissionRequest{
		UserID:   userID,
		Resource: resource,
		Action:   action,
	}
	data, err := c.do(http.MethodPost, "/check", req)
	if err != nil {
		return err
	}

	var resp api.CheckPermissionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	if resp.Allowed {
		fmt.Printf("Allowed (scope: %s)\n", resp.Scope)
	} else {
		fmt.Println("Denied")
	}
	return nil
}

func printUsage() {
	fmt.Println("RBAC Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  rbac-client role <subcommand>")
	fmt.Println("  rbac-client perm <subcommand>")
	fmt.Println("  rbac-client parent <subcommand>")
	fmt.Println("  rbac-client user-role <subcommand>")
	fmt.Println("  rbac-client check <user> <resource> <action>")
	fmt.Println()
	fmt.Println("Role commands:")
	fmt.Println("  role list")
	fmt.Println("  role create <id> <name> [description]")
	fmt.Println("  role get <id>")
	fmt.Println("  role update <id> <name> [description]")
	fmt.Println("  role delete <id>")
	fmt.Println()
	fmt.Println("Permission commands:")
	fmt.Println("  perm list <role_id>")
	fmt.Println("  perm add <role_id> <resource:action> <scope>")
	fmt.Println("  perm remove <role_id> <resource:action>")
	fmt.Println()
	fmt.Println("Parent commands:")
	fmt.Println("  parent list <role_id>")
	fmt.Println("  parent add <child_id> <parent_id>")
	fmt.Println("  parent remove <child_id> <parent_id>")
	fmt.Println()
	fmt.Println("User role commands:")
	fmt.Println("  user-role list <user_id>")
	fmt.Println("  user-role add <user_id> <role_id>")
	fmt.Println("  user-role remove <user_id> <role_id>")
	fmt.Println()
	fmt.Println("Check command:")
	fmt.Println("  check <user_id> <resource> <action>")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -server string")
	fmt.Println("        Server URL (default \"http://localhost:8080\")")
}

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "Server URL")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	client := NewClient(*serverURL)

	cmd := args[0]
	cmdArgs := args[1:]

	var err error

	switch cmd {
	case "role":
		if len(cmdArgs) == 0 {
			printUsage()
			os.Exit(1)
		}
		subCmd := cmdArgs[0]
		subArgs := cmdArgs[1:]
		switch subCmd {
		case "list":
			err = client.RoleList()
		case "create":
			if len(subArgs) < 2 {
				fmt.Println("Usage: role create <id> <name> [description]")
				os.Exit(1)
			}
			desc := ""
			if len(subArgs) >= 3 {
				desc = strings.Join(subArgs[2:], " ")
			}
			err = client.RoleCreate(subArgs[0], subArgs[1], desc)
		case "get":
			if len(subArgs) < 1 {
				fmt.Println("Usage: role get <id>")
				os.Exit(1)
			}
			err = client.RoleGet(subArgs[0])
		case "update":
			if len(subArgs) < 2 {
				fmt.Println("Usage: role update <id> <name> [description]")
				os.Exit(1)
			}
			desc := ""
			if len(subArgs) >= 3 {
				desc = strings.Join(subArgs[2:], " ")
			}
			err = client.RoleUpdate(subArgs[0], subArgs[1], desc)
		case "delete":
			if len(subArgs) < 1 {
				fmt.Println("Usage: role delete <id>")
				os.Exit(1)
			}
			err = client.RoleDelete(subArgs[0])
		default:
			printUsage()
			os.Exit(1)
		}

	case "perm":
		if len(cmdArgs) == 0 {
			printUsage()
			os.Exit(1)
		}
		subCmd := cmdArgs[0]
		subArgs := cmdArgs[1:]
		switch subCmd {
		case "list":
			if len(subArgs) < 1 {
				fmt.Println("Usage: perm list <role_id>")
				os.Exit(1)
			}
			err = client.PermList(subArgs[0])
		case "add":
			if len(subArgs) < 3 {
				fmt.Println("Usage: perm add <role_id> <resource:action> <scope>")
				os.Exit(1)
			}
			parts := strings.SplitN(subArgs[1], ":", 2)
			if len(parts) != 2 {
				fmt.Println("Invalid permission format, expected resource:action")
				os.Exit(1)
			}
			err = client.PermAdd(subArgs[0], parts[0], parts[1], subArgs[2])
		case "remove":
			if len(subArgs) < 2 {
				fmt.Println("Usage: perm remove <role_id> <resource:action>")
				os.Exit(1)
			}
			parts := strings.SplitN(subArgs[1], ":", 2)
			if len(parts) != 2 {
				fmt.Println("Invalid permission format, expected resource:action")
				os.Exit(1)
			}
			err = client.PermRemove(subArgs[0], parts[0], parts[1])
		default:
			printUsage()
			os.Exit(1)
		}

	case "parent":
		if len(cmdArgs) == 0 {
			printUsage()
			os.Exit(1)
		}
		subCmd := cmdArgs[0]
		subArgs := cmdArgs[1:]
		switch subCmd {
		case "list":
			if len(subArgs) < 1 {
				fmt.Println("Usage: parent list <role_id>")
				os.Exit(1)
			}
			err = client.ParentList(subArgs[0])
		case "add":
			if len(subArgs) < 2 {
				fmt.Println("Usage: parent add <child_id> <parent_id>")
				os.Exit(1)
			}
			err = client.ParentAdd(subArgs[0], subArgs[1])
		case "remove":
			if len(subArgs) < 2 {
				fmt.Println("Usage: parent remove <child_id> <parent_id>")
				os.Exit(1)
			}
			err = client.ParentRemove(subArgs[0], subArgs[1])
		default:
			printUsage()
			os.Exit(1)
		}

	case "user-role":
		if len(cmdArgs) == 0 {
			printUsage()
			os.Exit(1)
		}
		subCmd := cmdArgs[0]
		subArgs := cmdArgs[1:]
		switch subCmd {
		case "list":
			if len(subArgs) < 1 {
				fmt.Println("Usage: user-role list <user_id>")
				os.Exit(1)
			}
			err = client.UserRoleList(subArgs[0])
		case "add":
			if len(subArgs) < 2 {
				fmt.Println("Usage: user-role add <user_id> <role_id>")
				os.Exit(1)
			}
			err = client.UserRoleAdd(subArgs[0], subArgs[1])
		case "remove":
			if len(subArgs) < 2 {
				fmt.Println("Usage: user-role remove <user_id> <role_id>")
				os.Exit(1)
			}
			err = client.UserRoleRemove(subArgs[0], subArgs[1])
		default:
			printUsage()
			os.Exit(1)
		}

	case "check":
		if len(cmdArgs) < 3 {
			fmt.Println("Usage: check <user_id> <resource> <action>")
			os.Exit(1)
		}
		err = client.Check(cmdArgs[0], cmdArgs[1], cmdArgs[2])

	default:
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
