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

import {Component,ViewEncapsulation} from '@angular/core';
import { RouterModule,Router } from '@angular/router';
import {AF} from './authfire'

/**
 * LoginComponent for signing in with Github.
 */
@Component({
  selector: 'login',
  styleUrls: ['./login.component.scss'],
  templateUrl: './login.component.html',
})
export class LoginComponent {
  constructor(public afService: AF, private router: Router) {}
  login() {
    this.afService.loginWithGitHub().then((data) => {
      // Send them to the homepage if they are logged in
      this.router.navigate(['']);
    }).catch((error)=>{
        console.error(error);
    })
  }
}
