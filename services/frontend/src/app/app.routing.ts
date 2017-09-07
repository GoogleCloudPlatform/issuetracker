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

import { NgModule,ModuleWithProviders } from '@angular/core';
import { Routes,RouterModule } from '@angular/router';
import {AppComponent} from './app.component'
import {LoginComponent} from './login/login.component'
import { RepositoriesComponent } from './repositories/repositories.component';
import { PreferencesComponent } from './preferences/preferences.component';
import { NotificationsComponent } from './notifications/notifications.component';


export const APP_ROUTES: Routes = [
    {path: '', pathMatch: 'full', redirectTo:'repositories'},
    {path: 'login', component: LoginComponent},
    {path: 'repositories', component: RepositoriesComponent},
    {path: 'preferences', component: PreferencesComponent},
    {path: 'notifications', component: NotificationsComponent},
    {path: 'repos', component: RepositoriesComponent},
];

export const AuthenticatedRoutes = [
    {name: 'My Repositories', route: '/repositories',icon: 'dashboard'},
    {name: 'Notifications', route: '/notifications',icon: 'notifications'},
    {name: 'Preferences', route: '/preferences',icon: 'settings'},
]

export const AnonymousRoutes = [
    {name: 'Login', route: '/login',icon: 'person'},
]

export const appRoutingProviders: any[] = [
];

export const routing: ModuleWithProviders = RouterModule.forRoot(APP_ROUTES);