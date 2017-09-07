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
  MaterialModule,MdTableModule,MdPaginator,MdProgressSpinner
} from '@angular/material'
import {Component,OnInit,ElementRef, ViewChild,Injectable,NgModule} from '@angular/core';
import {DataSource,CdkTableModule} from '@angular/cdk';
import {BehaviorSubject} from 'rxjs/BehaviorSubject';
import {Observable} from 'rxjs/Observable';
import 'rxjs/add/operator/startWith';
import 'rxjs/add/observable/merge';
import 'rxjs/add/operator/map';
import 'rxjs/add/operator/debounceTime';
import 'rxjs/add/operator/distinctUntilChanged';
import 'rxjs/add/observable/fromEvent';
import {DialogsService} from './dialog.service';
import {UserDataService} from '../misc/services'
import {AF} from '../login/authfire'

var showSpinner = false;

@Component({
  selector: 'app-repositories',
  templateUrl: './repositories.component.html',
  styleUrls: ['./repositories.component.scss']
})
export class RepositoriesComponent implements OnInit {
  displayedColumns = ['repoName', 'openIssues', 'latestIssue'];
  repoDatabase = new RepoDatabase();
  dataSource: RepoDataSource | null;
  @ViewChild(MdPaginator) paginator: MdPaginator;
  @ViewChild('filter') filter: ElementRef;

  constructor (
    private afService: AF,
    private dialogsService: DialogsService
  ) {}

  ngOnInit() {
    this.dataSource = new RepoDataSource(this.repoDatabase, this.paginator);
    Observable.fromEvent(this.filter.nativeElement, 'keyup')
    .debounceTime(150)
    .distinctUntilChanged()
    .subscribe(() => {
      if (!this.dataSource) { return; }
      this.dataSource.filter = this.filter.nativeElement.value;
    });

    this.afService.user.subscribe(
      (user) => {
        if(user != null) {
          user.getIdToken().then(token=>{
            this.repoDatabase.populate(token)
          });
        }
      });
  }

  showLoader(){return showSpinner}

  public openAddDialog() {
    this.dialogsService
      .add()
      .subscribe((data)=>{
       if (data){
          showSpinner = true;
          this.afService.user.subscribe(
            (user) => {
              if(user != null) {
                user.getIdToken().then(token=>{
                  this.repoDatabase.subscribe(token,data)
                });
              }
            });
        }
      });
  }
  public openSettingsDialog(repo: string) {
    this.dialogsService
      .settings(repo)
      .subscribe((data)=>{
       if (data){
          this.afService.user.subscribe(
            (user) => {
              if(user != null) {
                user.getIdToken().then(token=>{
                    // Call Update subsciption here
                    this.repoDatabase.updatePreference(token,repo,data)
                });
              }
            });
        }
      });
  }
  public removeRepo(index: number,name: string) {
    this.afService.user.subscribe(
      (user) => {
        if(user != null) {
          user.getIdToken().then(token=>{
            this.repoDatabase.removeRepo(token,index,name)
          });
        }
      });

  }
}


export interface RepoData {
  index: number;
  name: string;
  openIssues: number;
  lastUpdate: number;
  url: string;
}

/** The database that the data source uses to retrieve data for the table. */
export class RepoDatabase {
  /** Stream that emits whenever the data has been modified. */
  dataChange: BehaviorSubject<RepoData[]> = new BehaviorSubject<RepoData[]>([]);
  get data(): RepoData[] { return this.dataChange.value; }

  constructor() {}

  populate(token: string)
  {
    showSpinner = true;
    var result = fetch('/api/users/repos',{
      method:'GET',
      headers:{
        'Authorization': token
      }
    }).then((response)=>{
        showSpinner = false;
        return response.json()
    },(err)=>{console.log(err); showSpinner = false;}).then((json)=>{
        // Populate RepoDatabase
        json.forEach(repo => {
          if(this.exists(repo["Name"])==false){
            this.addRepo(repo["Name"],repo["IssuesOpen"],repo["UpdatedAt"])
        }
      });
      showSpinner = false;
    }).catch((err)=>{console.log(err); showSpinner = false;});
  }

  subscribe(token: string, data)
  {
    var formData = new FormData()
    formData.append("repo",data["repo"]);
    // Make an API call to subscribe to the repository
    var result = fetch('/api/subscriptions/add',{
      method:'POST',
      headers:{
        'Authorization': token
      },
      body:formData
    }).then((response)=>{
        return response.json()
    },(err)=>{console.log(err)}).then((json)=>{
      if(!json["Code"]){
        this.populate(token)
      }
    });

  }
  updatePreference(token: string,repo: string, data)
  {
    var formData = new FormData();
    formData.append("repo",repo);
    formData.append("defaultEmail",data["email"]);
    formData.append("settings",JSON.stringify(data["settings"]));
    // Make an API call to update notification preferences
    var result = fetch('/api/subscriptions/update',{
      method:'POST',
      headers:{
        'Authorization': token
      },
      body:formData
    }).then((response)=>{
        return response.json();
    },(err)=>{console.log(err)}).then((json)=>{
      if(!json["Code"]){
        this.populate(token);
      }
    });
  }

  /** Adds a new repo to the database. */
  addRepo(repo: string,open: number,lastUpdate: string) {
    const copiedData = this.data.slice();
    copiedData.push(this.createNewRepo(repo,open,lastUpdate));
    this.dataChange.next(copiedData);
  }

  /** Removes repo from the database. */
  removeRepo(token: string,repoIndex: number, repoName: string) {
    let pos = this.data.findIndex((value,index,obj)=>{
      if(value.index == repoIndex)
        return true;
      return false;
    })
    if (pos != -1){
      var formData = new FormData();
      formData.append("repo",repoName);
      // Make an API call to subscribe to the repository
      var result = fetch('/api/subscriptions/remove',{
        method:'POST',
        headers:{
          'Authorization': token
        },
        body:formData
      }).then((response)=>{
        return response.json();
      },(err)=>{console.log(err)}).then((json)=>{
        if(json["Code"] == 200){
          const copiedData = this.data.slice();
          copiedData.splice(pos,1);
          this.dataChange.next(copiedData)
        }
      });
    }

  }

  exists(repo: string){
   let pos =  this.data.findIndex((value,index,obj)=>{
      if(value.name == repo)
        return true;
      return false;
    })
    if (pos == -1)
      return false;
    else{
      return true;
    }

  }
  /** Builds and returns a new repo. */
  private createNewRepo(repo: string,open: number,lastUpdate: string) {
    var currentDate = new Date();
    var updateDate = new Date(lastUpdate);
    // get the difference between the dates in milliseconds and divide by 3600000
    var difference=Math.abs(updateDate.getTime()-currentDate.getTime()) / 36e5;
    difference = Math.trunc(difference)
    return {
      index: (this.data.length + 1),
      name: repo,
      openIssues: open,
      lastUpdate: difference,
      url: "https://github.com/"+repo+"/issues"
    };
  }
}

/**
* Data source to provide what data should be rendered in the table. Note that the data source
* can retrieve its data in any way. In this case, the data source is provided a reference
* to a common data base, RepoDatabase. It is not the data source's responsibility to manage
* the underlying data. Instead, it only needs to take the data and send the table exactly what
* should be rendered.
*/
export class RepoDataSource extends DataSource<any> {
  _filterChange = new BehaviorSubject('');
  get filter(): string { return this._filterChange.value; }
  set filter(filter: string) { this._filterChange.next(filter); }

  constructor(private _repoDatabase: RepoDatabase, private _paginator: MdPaginator) {
    super();
  }

  /** Connect function called by the table to retrieve one stream containing the data to render. */
  connect(): Observable<RepoData[]> {
    const displayDataChanges = [
      this._repoDatabase.dataChange,
      this._paginator.page,
      this._filterChange,
    ];

    return Observable.merge(...displayDataChanges).map(() => {
      const data = this._repoDatabase.data.slice().filter((item: RepoData) => {
        let searchStr = (item.name).toLowerCase();
        return searchStr.indexOf(this.filter.toLowerCase()) != -1;
      });

      // Grab the page's slice of data.
      const startIndex = this._paginator.pageIndex * this._paginator.pageSize;
      return data.splice(startIndex, this._paginator.pageSize);
    });
  }

  disconnect() {}
}


