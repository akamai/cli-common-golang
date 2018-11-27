/*
 * Copyright 2018. Akamai Technologies, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package akamai_cli

import (
	"fmt"
	"os"
	"time"

	spnr "github.com/briandowns/spinner"
	"github.com/fatih/color"
)

func StartSpinner(prefix string, finalMsg string) *spnr.Spinner {
	spinner = spnr.New(spnr.CharSets[26], 500*time.Millisecond)
	spinner.Writer = App.ErrWriter
	spinner.Prefix = prefix
	spinner.FinalMSG = finalMsg
	if log := os.Getenv("AKAMAI_LOG"); len(log) > 0 || !isTTY() {
		fmt.Println(prefix)
	} else {
		spinner.Start()
	}

	return spinner
}

func StopSpinner(finalMsg string, usePrefix bool) {
	if spinner == nil {
		return
	}
	if usePrefix {
		spinner.FinalMSG = spinner.Prefix + finalMsg
	} else {
		spinner.FinalMSG = finalMsg
	}

	if log := os.Getenv("AKAMAI_LOG"); len(log) > 0 || !isTTY() {
		fmt.Println(spinner.FinalMSG)
		return
	}
	spinner.Stop()
}

func StopSpinnerOk() {
	StopSpinner(fmt.Sprintf("... [%s]\n", color.GreenString("OK")), true)
}

func StopSpinnerWarnOk() {
	StopSpinner(fmt.Sprintf("... [%s]\n", color.CyanString("OK")), true)
}

func StopSpinnerWarn() {
	StopSpinner(fmt.Sprintf("... [%s]\n", color.CyanString("WARN")), true)
}

func StopSpinnerFail() {
	StopSpinner(fmt.Sprintf("... [%s]\n", color.RedString("FAIL")), true)
}
