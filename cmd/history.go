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
	"encoding/json"
	"os"
	"regexp"

	"github.com/k1LoW/slackln/client"
	"github.com/slack-go/slack"
	"github.com/spf13/cobra"
)

var tsFieldRe = regexp.MustCompile(`"ts":"([0-9.]+)"`)

var (
	channel  string
	duration string
	latest   string
	oldest   string
	raw      bool
)

// historyCmd represents the history command
var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "print channel history",
	Long:  `print channel history.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		c, err := client.New(getToken())
		if err != nil {
			return err
		}

		msgChan := make(chan slack.Msg)

		go func() {
			err := c.GetHistory(ctx, msgChan, channel, duration, latest, oldest)
			if err != nil {
				cmd.PrintErrln(err)
				cancel()
				os.Exit(1)
			}
		}()

		for m := range msgChan {
			if !raw {
				m.User = c.GetUserNameByID(ctx, m.User)
				m.Text = c.HumanizeMessage(ctx, m.Text)
				for i, a := range m.Attachments {
					m.Attachments[i].Text = c.HumanizeMessage(ctx, a.Text)
					for j, f := range a.Fields {
						m.Attachments[i].Fields[j].Value = c.HumanizeMessage(ctx, f.Value)
					}
				}
				for i, r := range m.Reactions {
					for j, u := range r.Users {
						m.Reactions[i].Users[j] = c.GetUserNameByID(ctx, u)
					}
				}
			}
			b, err := json.Marshal(m)
			if err != nil {
				cancel()
				return err
			}
			s := string(b)
			if !raw {
				s = tsFieldRe.ReplaceAllStringFunc(s, func(in string) string {
					raw := tsFieldRe.ReplaceAllString(in, "$1")
					pl, err := c.CreateParmalink(ctx, channel, raw)
					if err != nil {
						cmd.PrintErrln(err)
					}
					t := c.HumanizeTimestamp(raw)
					out := `"ts":"` + t + `","ts_raw":"` + raw + `","permalink":"` + pl + `"`
					return out
				})
			}
			cmd.Println(s)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)
	historyCmd.Flags().StringVarP(&channel, "channel", "c", "", "Slack channel")
	historyCmd.Flags().StringVarP(&duration, "duration", "", "1day", "duration of time range of messages.")
	historyCmd.Flags().StringVarP(&latest, "latest", "", "", "end of time range of messages.")
	historyCmd.Flags().StringVarP(&oldest, "oldest", "", "", "start of time range of messages.")
	historyCmd.Flags().BoolVarP(&raw, "raw", "", false, "print raw messages.")
}
