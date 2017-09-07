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

import { Component,OnInit, ViewChild} from '@angular/core';
import {AF} from '../login/authfire'
import {NgForm} from '@angular/forms'
import {MdSnackBar,MdInputModule} from '@angular/material';
import { Observable } from 'rxjs/Observable';
import {UserDataService} from '../misc/services'

@Component({
  selector: 'app-preferences',
  templateUrl: './preferences.component.html',
  styleUrls: ['./preferences.component.scss']
})
export class PreferencesComponent implements OnInit {


  defaultEmail: string
  userData: JSON

  @ViewChild('preferenceForm') prefForm: NgForm;

  constructor(
      public snackBar: MdSnackBar,
      private afService: AF,
      private userEmitter:UserDataService,
    ){}

  ngOnInit() {
     this.userEmitter.getData().subscribe(data => this.setUserData(data));
  }

  setUserData(data){
    this.userData = data
    this.defaultEmail = data["Email"]
  }

  onSubmit(){
    this.afService.user.subscribe((user)=>
    {
      user.getIdToken().then(token=>{
        var formData = new FormData()
        formData.append("email",this.defaultEmail)
        var result = fetch('/api/users/update',{
          method:'POST',
          headers:{
            'Authorization': token
          },
          body:formData
        }).then((response)=>{
            return response.json()
        },(err)=>{console.log(err)}).then((json)=>{
          if(!json["Code"]){
            this.snackBar.open("Updated Preferences", "close", {
              duration: 2000,
            });
            this.userData["Email"] = this.defaultEmail
            this.userEmitter.saveData(this.userData)
            this.prefForm.form.markAsPristine()
          }
        });
      });
    });
  }

}