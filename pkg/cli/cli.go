package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	ucli "github.com/urfave/cli"

	"github.com/perryrh0dan/passline/pkg/core"
	"github.com/perryrh0dan/passline/pkg/renderer"
	"github.com/perryrh0dan/passline/pkg/storage"
)

var passline *core.Core

func Init(ctx context.Context) {
	var err error
	passline, err = core.NewCore(ctx)
	if err != nil {
		renderer.CoreInstanceError()
		os.Exit(1)
	}
}

func CreateBackup(ctx context.Context, c *ucli.Context) error {
	args := c.Args()
	renderer.BackupMessage()

	// User input path
	path, err := argOrInput(args, 0, "Path", "")
	if err != nil {
		return err
	}

	err = passline.CreateBackup(ctx, path)
	if err != nil {
		return err
	}

	renderer.SuccessfulCreatedBackup(path)
	return nil
}

func CreateItem(ctx context.Context, c *ucli.Context) error {
	args := c.Args()
	renderer.CreateMessage()

	// User input name
	name, err := argOrInput(args, 0, "URL", "")
	handle(err)

	// User input username
	username, err := argOrInput(args, 1, "Username/Login", "")
	handle(err)

	// Check if name, username combination exists
	exists, err := passline.Exists(ctx, name, username)
	if err != nil {
		return err
	}

	if exists {
		renderer.NameUsernameAlreadyExists()
		return nil
	}

	password, err := Input("Password", "")
	if err != nil {
		return err
	}

	globalPassword := getPassword("Enter Global Password: ")
	println()

	credential, err := passline.CreateItem(ctx, name, username, password, globalPassword)
	if err != nil {
		return err
	}

	renderer.DisplayCredential(credential)
	return nil
}

func DeleteItem(ctx context.Context, c *ucli.Context) error {
	// Get all Sites
	names, err := passline.GetSiteNames(ctx)
	if err != nil {
		return err
	}

	// Check if site exists
	if len(names) <= 0 {
		renderer.NoItemsMessage()
		return nil
	}

	args := c.Args()
	renderer.DeleteMessage()

	item, err := selectItem(ctx, args, names)
	if err != nil {
		return err
	}

	credential, err := selectCredential(args, item)
	if err != nil {
		return err
	}

	err = passline.DeleteItem(ctx, item.Name, credential.Username)
	if err != nil {
		return err
	}

	renderer.SuccessfulDeletedItem(item.Name, credential.Username)
	return nil
}

func DisplayItem(ctx context.Context, c *ucli.Context) error {
	// Get all Sites
	names, err := passline.GetSiteNames(ctx)
	if err != nil {
		return err
	}

	// Check if site exists
	if len(names) <= 0 {
		renderer.NoItemsMessage()
		os.Exit(0)
	}

	args := c.Args()
	renderer.DisplayMessage()

	item, err := selectItem(ctx, args, names)
	handle(err)

	credential, err := selectCredential(args, item)
	handle(err)

	// Get global password.
	globalPassword := getPassword("Enter Global Password: ")
	println()

	// Check global password.
	valid, err := passline.CheckPassword(ctx, globalPassword)
	if err != nil || !valid {
		handle(err)
	}

	err = passline.DecryptPassword(&credential, globalPassword)
	if err != nil {
		return err
	}

	renderer.DisplayCredential(credential)
	return nil
}

func EditItem(ctx context.Context, c *ucli.Context) error {
	// Get all Sites
	names, err := passline.GetSiteNames(ctx)
	if err != nil {
		return err
	}

	// Check if site exists
	if len(names) <= 0 {
		renderer.NoItemsMessage()
		os.Exit(0)
	}

	args := c.Args()
	renderer.DeleteMessage()

	item, err := selectItem(ctx, args, names)
	handle(err)

	credential, err := selectCredential(args, item)
	handle(err)

	// Get new username
	newUsername, err := Input("Please enter a new Username/Login []: (%s) ", credential.Username)
	handle(err)

	if newUsername == "" {
		newUsername = credential.Username
	}

	err = passline.EditItem(ctx, item.Name, credential.Username, newUsername)
	handle(err)

	renderer.SuccessfulChangedItem(item.Name, credential.Username)

	return nil
}

func GenerateItem(ctx context.Context, c *ucli.Context) error {
	args := c.Args()
	renderer.GenerateMessage()

	// User input name
	name, err := argOrInput(args, 0, "URL", "")
	handle(err)

	// User input username
	username, err := argOrInput(args, 1, "Username/Login", "")
	handle(err)

	// Check if name, username combination exists
	exists, err := passline.Exists(ctx, name, username)
	if exists {
		os.Exit(0)
	}

	globalPassword := getPassword("Enter Global Password: ")
	println()

	credential, err := passline.GenerateItem(ctx, name, username, globalPassword)
	handle(err)

	err = clipboard.WriteAll(credential.Password)
	if err != nil {
		renderer.ClipboardError()
		os.Exit(0)
	}

	renderer.SuccessfulCopiedToClipboard(name, credential.Username)
	return nil
}

func ListItems(ctx context.Context, c *ucli.Context) error {
	args := c.Args()

	if len(args) >= 1 {
		item, err := passline.GetSite(ctx, args[0])
		if err != nil {
			renderer.InvalidName(args[0])
			os.Exit(0)
		}

		renderer.DisplayItem(item)
	} else {
		items, err := passline.GetSites(ctx)
		if err != nil {
			return nil
		}

		if len(items) == 0 {
			renderer.NoItemsMessage()
		}
		renderer.DisplayItems(items)
	}

	return nil
}

func RestoreBackup(ctx context.Context, c *ucli.Context) error {
	args := c.Args()
	renderer.RestoreMessage()

	// User input path
	path, err := argOrInput(args, 0, "Path", "")
	if err != nil {
		return err
	}

	err = passline.RestoreBackup(ctx, path)
	if err != nil {
		return err
	}

	renderer.SuccessfulRestoredBackup(path)
	return nil
}

func argOrInput(args []string, index int, message string, value string) (string, error) {
	input := ""
	if len(args)-1 >= index {
		input = args[index]
	}
	if input == "" {
		message := fmt.Sprintf("Please enter a %s []: ", message)
		var err error
		input, err = Input(message, value)
		if err != nil {
			return "", err
		}
	}

	return input, nil
}

func argOrSelect(args []string, index int, message string, items []string) (string, error) {
	input := ""
	if len(args)-1 >= index {
		input = args[index]
	}
	if input == "" {
		if len(items) > 1 {
			message := fmt.Sprintf("Please select a %s: ", message)
			selection, err := Select(message, items)
			if err != nil {
				return "", err
			}

			input = items[selection]
			fmt.Printf("%s%s\n", message, input)
		} else if len(items) == 1 {
			fmt.Printf("Selected %s: %s\n", message, items[0])
			return items[0], nil
		}
	}

	return input, nil
}

func handle(err error) {
	if err != nil {
		os.Exit(1)
	}
}

func selectItem(ctx context.Context, args, names []string) (storage.Item, error) {
	name, err := argOrSelect(args, 0, "URL", names)
	handle(err)

	// Get item
	item, err := passline.GetSite(ctx, name)
	if err != nil {
		renderer.InvalidName(name)
		os.Exit(0)
	}

	return item, nil
}

func selectCredential(args []string, item storage.Item) (storage.Credential, error) {
	username, err := argOrSelect(args, 1, "Username/Login", item.GetUsernameArray())
	handle(err)

	// Check if name, username combination exists
	credential, err := item.GetCredentialByUsername(username)
	if err != nil {
		renderer.InvalidUsername(item.Name, username)
		os.Exit(0)
	}

	return credential, nil
}
