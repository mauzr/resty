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

package color

var (
	// Off turns the pixel off.
	Off = RGBW{}
	// Bright sets the pixel to as bright as possible.
	Bright = RGBW{White: 1.0}
	// Unhealthy indicates that something is wrong.
	Unhealthy = RGBW{Red: 1.0}
	// Unmanaged indicates that the pixel is not actively managed.
	Unmanaged = Unhealthy
	// Alert indicates that something needs user attention.
	Alert = Unhealthy
	// Healthy indicated that everything is fine.
	Healthy = RGBW{Green: 1.0}
)
