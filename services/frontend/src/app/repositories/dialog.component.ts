// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import {
  MdDialogRef,
  MdInputContainer,
  MaterialModule,
  MdInputModule,
  MdOptionModule,
  MdSelectModule,
} from '@angular/material';
import { Component,OnInit} from '@angular/core';
import {UserDataService} from '../misc/services'

@Component({
    selector: 'add-dialog',
    templateUrl: './add.dialog.html',
  })
  export class AddDialog {
    constructor(public dialogRef: MdDialogRef<AddDialog>) {
    }
    public addRepo(url: string){
      if (url)
      {
        var data = {
          "repo":url,
        };
        return this.dialogRef.close(data)

      }
      return this.dialogRef.close()
    }
  }

  export class Settings {
    constructor(
      public IssueOpen:   number,
      public IssueClose:  number,
      public IssueReopen: number,
      public NewComment:  number,
      public NoComment:   number,
    ){}
  }


  @Component({
    selector: 'settings-dialog',
    templateUrl:'./settings.dialog.html',
    styleUrls: ['./dailog.component.scss']
  })
  export class SettingsDialog implements OnInit {

    // Frequency
    frequency = [
      {view:"Never",value:1},
      {view:"Daily",value:2},
      {view:"Weekly",value:3},
      {view:"Monthly",value:4},
    ]
    // Default preferences
    settings = new Settings(2,2,2,2,2)
    defaultEmail = ""
    repo: string
    private storedPreference:any
    private updatedData:any
    private repoIndex: number

    constructor(
      public dialogRef: MdDialogRef<SettingsDialog>,
      private userEmitter:UserDataService
    ){}

    ngOnInit() {
      this.userEmitter.getData()
      .subscribe(data => {
        this.defaultEmail = data["DefaultEmail"]
        var subs = data["Subscriptions"]
        if (subs){
          subs.forEach((element,index) => {
            if(element["Repo"] == this.repo){
                this.repoIndex = index
                this.defaultEmail = element["DefaultEmail"]
                this.settingsFromJSON(element["EmailPreference"])
                this.storedPreference = element["EmailPreference"]
                this.updatedData = data
            }
          });
        }
      });
    }

    settingsFromJSON(data) {
      this.settings = new Settings(
          data["IssueOpen"],
          data["IssueClose"],
          data["IssueReopen"],
          data["NewComment"],
          data["NoComment"]
        )
    }

    updateStoredPreference(){
      let data = this.settings
      this.storedPreference["IssueOpen"] = data["IssueOpen"];
      this.storedPreference["IssueClose"] = data["IssueClose"];
      this.storedPreference["IssueReopen"] = data["IssueReopen"];
      this.storedPreference["NewComment"] = data["NewComment"];
      this.storedPreference["NoComment"] = data["NoComment"];
    }

    onSubmit() {
      console.log(this.updatedData)
      var data = {
        "settings": this.settings,
        "email":this.defaultEmail
      };
      this.updateStoredPreference()
      this.updatedData["Subscriptions"][this.repoIndex]["EmailPreference"] = this.storedPreference
      this.updatedData["Subscriptions"][this.repoIndex]["DefaultEmail"] = this.defaultEmail
      this.userEmitter.saveData(this.updatedData)

      return this.dialogRef.close(data)
    }
  }