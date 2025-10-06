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
	"io"
	"os"

	"github.com/spf13/cobra"
)

// configViewCmd represents the configView command
var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "View the configuration file",
	RunE: func(_ *cobra.Command, _ []string) error {
		file, err := os.Open(configPath)
		if err != nil {
			return err
		}

		var closeError error
		defer func() {
			closeError = file.Close()
		}()

		if _, err := io.Copy(os.Stdout, file); err != nil {
			return err
		}

		return closeError
	},
}

func init() {
	configCmd.AddCommand(configViewCmd)
}
