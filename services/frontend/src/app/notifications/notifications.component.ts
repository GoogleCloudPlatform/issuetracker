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
import {UserDataService} from '../misc/services'
import {AF} from '../login/authfire'

var showSpinner = false;

@Component({
  selector: 'app-notifications',
  templateUrl: './notifications.component.html',
  styleUrls: ['./notifications.component.scss']
})
export class NotificationsComponent implements OnInit {
  displayedColumns = ['notification'];
  notificationDatabase = new NotificationDatabase();
  dataSource: NotificationDataSource | null;
  @ViewChild(MdPaginator) paginator: MdPaginator;
  @ViewChild('filter') filter: ElementRef;

  constructor (
    private afService: AF,
  ) {}

  ngOnInit() {
    this.dataSource = new NotificationDataSource(this.notificationDatabase, this.paginator);
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
            this.notificationDatabase.populate(token)
          });
        }
      });
  }

  showLoader(){return showSpinner}

}


export interface NotificationData {
  index: number;
  id: number;
  userID: number;
  email: string;
  type: string;
  sentAt: string;
}

/** The database that the data source uses to retrieve data for the table. */
export class NotificationDatabase {
  /** Stream that emits whenever the data has been modified. */
  dataChange: BehaviorSubject<NotificationData[]> = new BehaviorSubject<NotificationData[]>([]);
  get data(): NotificationData[] { return this.dataChange.value; }

  constructor() {}

  populate(token: string)
  {
    showSpinner = true;
    var result = fetch('/api/notifications',{
      method:'GET',
      headers:{
        'Authorization': token
      }
    }).then((response)=>{
        showSpinner = false;
        return response.json()
    },(err)=>{console.log(err); showSpinner = false;}).then((json)=>{
        // Populate NotificationDatabase
        json.forEach(notification => {
          if(this.exists(notification["ID"])==false){
            this.addNotification(
              notification["ID"],
              notification["UserID"],
              notification["Email"],
              notification["Type"],
              notification["CreatedAt"])
        }
      });
      showSpinner = false;
    }).catch((err)=>{console.log(err); showSpinner = false;});
  }


  /** Adds a new notification to the database. */
  addNotification(id,userID,email,type:number,createdAt) {
    const copiedData = this.data.slice();
    var notification = this.createNewNotification(id,userID,email,type,createdAt)
    copiedData.push(notification);
    this.dataChange.next(copiedData);
  }

  exists(notification: number){
   let pos =  this.data.findIndex((value,index,obj)=>{
      if(value.id == notification)
        return true;
      return false;
    })
    if (pos == -1)
      return false;
    else{
      return true;
    }

  }
  /** Builds and returns a new notification. */
  private createNewNotification(id,userID,email,type:number,createdAt) {

    var sent = new Date(createdAt);
    var kind =""
    switch(type){
      case 2: {
        kind = "daily"
        break;
      }
      case 3: {
        kind = "weekly"
        break;
      }
      case 4: {
        kind = "monthly"
        break;
      }
    }
    let options = {
        weekday: 'long',
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    };
    return {
      index: (this.data.length + 1),
      id: id,
      userID: userID,
      email: email,
      type: kind,
      sentAt: sent.toLocaleDateString('en-us',options)
    };
  }
}

/**
* Data source to provide what data should be rendered in the table. Note that the data source
* can retrieve its data in any way. In this case, the data source is provided a reference
* to a common data base, NotificationDatabase. It is not the data source's responsibility to manage
* the underlying data. Instead, it only needs to take the data and send the table exactly what
* should be rendered.
*/
export class NotificationDataSource extends DataSource<any> {
  _filterChange = new BehaviorSubject('');
  get filter(): string { return this._filterChange.value; }
  set filter(filter: string) { this._filterChange.next(filter); }

  constructor(private _notificationDatabase: NotificationDatabase, private _paginator: MdPaginator) {
    super();
  }

  /** Connect function called by the table to retrieve one stream containing the data to render. */
  connect(): Observable<NotificationData[]> {
    const displayDataChanges = [
      this._notificationDatabase.dataChange,
      this._paginator.page,
      this._filterChange,
    ];

    return Observable.merge(...displayDataChanges).map(() => {
      const data = this._notificationDatabase.data.slice().filter((item: NotificationData) => {
        let searchStr = item.type+item.email+item.sentAt;
        return searchStr.indexOf(this.filter) != -1;
      });

      // Grab the page's slice of data.
      const startIndex = this._paginator.pageIndex * this._paginator.pageSize;
      return data.splice(startIndex, this._paginator.pageSize);
    });
  }

  disconnect() {}
}


