/*
Copyright 2019 Alexander Sowitzki.

GNU Affero General Public License version 3 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://opensource.org/licenses/AGPL-3.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"go.eqrx.net/mauzr/pkg"
	"go.eqrx.net/mauzr/pkg/program"
)

func main() {
	p := program.NewProgram()
	pkg.SetupCommands(p)
	defer p.Wg.Wait()
	defer p.Cancel()
	if err := p.RootCommand.Execute(); err != nil {
		panic(err)
	}
}
