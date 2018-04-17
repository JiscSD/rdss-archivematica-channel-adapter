package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use: "validate",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validate(); err != nil {
			fmt.Println(err)
		}
		return nil
	},
}

var filename string

func init() {
	validateCmd.Flags().StringVarP(&filename, "file", "f", "", "")

	RootCmd.AddCommand(validateCmd)
}

func validate() error {
	if filename == "" {
		return errors.New("parameter empty")
	}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	msg := &message.Message{}
	err = json.Unmarshal(data, msg)
	if err != nil {
		return err
	}
	fmt.Printf("Message %s found!\n", msg.ID())
	validator, err := message.NewValidator()
	if err != nil {
		return err
	}
	result, err := validator.Validate(msg)
	if err != nil {
		return err
	}
	if !result.Valid() {
		fmt.Println("The message is invalid!")
		for _, issue := range result.Errors() {
			fmt.Println(issue)
		}
	}
	return nil
}
