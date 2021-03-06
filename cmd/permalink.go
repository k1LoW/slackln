/*
Copyright © 2020 Ken'ichiro Oyama <k1lowxb@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"context"

	"github.com/k1LoW/slackln/client"
	"github.com/spf13/cobra"
)

var messageTs string

// permalinkCmd represents the permalink command
var permalinkCmd = &cobra.Command{
	Use:   "permalink",
	Short: "print permalink URL of message",
	Long:  `print permalink URL of message.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		c, err := client.New(getToken())
		if err != nil {
			return err
		}
		l, err := c.GetPermalink(ctx, channel, messageTs)
		if err != nil {
			return err
		}
		cmd.Println(l)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(permalinkCmd)
	permalinkCmd.Flags().StringVarP(&channel, "channel", "c", "", "Slack channel")
	permalinkCmd.Flags().StringVarP(&messageTs, "ts", "t", "", "message timestamp")
}
