import { Component, OnInit } from '@angular/core';
import { Observable } from 'rxjs/Rx';
import websocketConnect from 'rxjs-websockets'
import { QueueingSubject } from 'queueing-subject'
import { Subscription } from 'rxjs/Subscription'
import 'rxjs/add/operator/scan';


class Message {
  nick: string
  content: string
}


@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements OnInit {

  messages: Observable<Message>
  sendQueue: QueueingSubject<any>
  subscription: Subscription
  msg = ''

  onSendClicked() {
    let message = new Message()
    message.nick = "Joel"
    message.content = this.msg

    this.sendQueue.next(message)
    this.msg = ''
  }

  ngOnInit() {
    this.sendQueue = new QueueingSubject<any>()
    let url = 'ws://127.0.0.1:8080/ws'
    
    let msgs = websocketConnect(url, this.sendQueue).messages

    let dummyMessage = new Message()
    dummyMessage.nick = "Joel"

    this.sendQueue.next(dummyMessage)                   
            
    this.messages = msgs.scan( (acc: Array<Message>, m, index) => {
      let msg = new Message()
      msg.nick = m.Nick
      msg.content = m.Content
      acc.push(msg)
      return acc
    },  [])
  }
}
