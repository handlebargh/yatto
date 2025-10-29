// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"github.com/spf13/cobra"
)

var (
	// ErrNoEditorSet is returned when the EDITOR environment variable is empty.
	ErrNoEditorSet = fmt.Errorf("environment variable EDITOR not set")

	// ErrInvalidEditorSet is returned when the EDITOR environment variable contains illegal characters.
	ErrInvalidEditorSet = fmt.Errorf("environment variable EDITOR contains illegal characters")

	// editorRegexp validates the EDITOR environement variable.
	editorRegexp = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// configEditCmd represents the config edit command
var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit the configuration file",
	RunE: func(_ *cobra.Command, _ []string) error {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			return ErrNoEditorSet
		} else if !editorRegexp.MatchString(editor) {
			return ErrInvalidEditorSet
		}

		cmd := exec.Command(editor, configPath) // #nosec G204 Command uses validated variables
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		return cmd.Run()
	},
}

func init() {
	configCmd.AddCommand(configEditCmd)
}
