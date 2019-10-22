package app

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/JiscSD/rdss-archivematica-channel-adapter/broker/message"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var file string

func NewCmdValidate(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate RDSS API JSON documents",
		RunE: func(cmd *cobra.Command, args []string) error {
			return doValidate(out)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "File")

	return cmd
}

func doValidate(out io.Writer) error {
	if file == "" {
		return errors.New("parameter empty")
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return errors.Wrap(err, "cannot read file")
	}
	msg := &message.Message{}
	err = json.Unmarshal(data, msg)
	if err != nil {
		return err
	}
	fmt.Fprintln(out, "Message found!", msg.ID())
	validator, err := message.NewValidator("strict")
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
			fmt.Fprintln(out, issue)
		}
	}
	return nil
}
