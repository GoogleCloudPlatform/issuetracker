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

import { BrowserModule } from '@angular/platform-browser';
import {BrowserAnimationsModule} from '@angular/platform-browser/animations';
import {CdkTableModule} from '@angular/cdk'
import { NgModule } from '@angular/core';
import { FormsModule }   from '@angular/forms';
import {
  MaterialModule,
  MdTableModule,
  MdInputModule,
  MdSnackBarModule,
  MdCardModule,
  MdProgressSpinnerModule,
  FullscreenOverlayContainer,
  OverlayContainer,
} from '@angular/material';
import 'hammerjs';
import { routing, appRoutingProviders } from './app.routing';
import {RouterModule} from '@angular/router';
import { AppComponent } from './app.component';
import { LoginComponent } from './login/login.component'
import { Angular2FontawesomeModule } from 'angular2-fontawesome/angular2-fontawesome'
import { RepositoriesComponent } from './repositories/repositories.component';
import { PreferencesComponent } from './preferences/preferences.component';
import { NotificationsComponent } from './notifications/notifications.component';
import { UserAvatar } from './misc/components';
import { AngularFireModule } from 'angularfire2';
import {environment} from "../environments/environment"
import {AF} from "./login/authfire"
import { AngularFireAuth } from 'angularfire2/auth';
import { DialogsModule } from "./repositories/dialog.module";
import {UserDataService} from './misc/services'


@NgModule({
  declarations: [
    AppComponent,
    LoginComponent,
    RepositoriesComponent,
    PreferencesComponent,
    NotificationsComponent,
    UserAvatar
  ],
  imports: [
    BrowserModule,
    routing,
    BrowserAnimationsModule,
    MaterialModule,
    FormsModule,
    MdTableModule,
    MdSnackBarModule,
    MdInputModule,
    CdkTableModule,
    MdProgressSpinnerModule,
    Angular2FontawesomeModule,
    AngularFireModule.initializeApp(environment.firebase),
    DialogsModule
  ],
  providers: [
    {provide: OverlayContainer, useClass: FullscreenOverlayContainer},
    appRoutingProviders,
    AngularFireAuth,
    AF,
    UserDataService,
  ],
  bootstrap: [AppComponent],
  exports: [RouterModule,RepositoriesComponent, PreferencesComponent, NotificationsComponent]
})
export class AppModule { }
