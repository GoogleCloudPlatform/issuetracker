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

import {Component,OnInit,ViewChild,ViewEncapsulation} from '@angular/core';
import { Routes, RouterModule,Router } from '@angular/router';
import {OverlayContainer,MdSidenav} from '@angular/material';
import {AuthenticatedRoutes,AnonymousRoutes} from './app.routing'
import {AF} from './login/authfire'
import { Subject } from 'rxjs/Subject';
import * as firebase from 'firebase/app';
import {UserDataService} from './misc/services'



@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss'],
  encapsulation: ViewEncapsulation.None,
})
export class AppComponent implements OnInit{

  public isLoggedIn: boolean;
  user: firebase.User;
  userData: JSON
  navItems = [];
  @ViewChild("start")
  private start: MdSidenav;

  // Setup services for Angular Fire
  constructor(
      private afService: AF,
      private router: Router,
      private userEmitter: UserDataService
  ) {}

  // Check the auth status for the user and handle redirects to login page
  ngOnInit(){
    this.authCheck();
    this.start.open()
  }

  authCheck(){
    // authCheck asynchronously checks if our user is logged it and will automatically
    // redirect them to the Login page when the status changes.
    this.afService.user.subscribe(
      (user) => {
        if(user == null) {
          this.router.navigate(['login']);
          this.isLoggedIn = false;
          this.navItems = AnonymousRoutes;
        }
        else {
          this.navItems = AuthenticatedRoutes;
          this.user = user;
          user.getIdToken().then(token=>{
            this.verifyToken(token)
          });
          this.isLoggedIn = true;
        }
      }
    );
  }

  private verifyToken(token: string)
  {
    var result = fetch('/api/auth',{
      method:'GET',
      headers:{
        'Authorization': token
      }
    }).then((response)=>{
        return response.json()
    },(err)=>{console.log(err)}).then((json)=>{
        if (json["Valid"]==true){
          this.userData = json["User"]
          this.userData["token"] = token
          this.userEmitter.saveData(this.userData)
          if(json["Onboard"]==true){
            this.router.navigate(['preferences']);
          }
          else if(this.isLoggedIn != true) {
            this.router.navigate(['']);
          }
        }
        else{
          this.logout()
        }
    }).catch((err)=>{console.log(err)});
  }

  logout() {
    this.afService.logout().catch((error)=>{
      console.error("Error while trying to sign out user:"+error)
      this.isLoggedIn=false
      this.router.navigate(['login']);
    });
  }
}
