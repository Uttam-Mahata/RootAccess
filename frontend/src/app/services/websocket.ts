import { Injectable, OnDestroy } from '@angular/core';
import { Subject, Observable } from 'rxjs';
import { filter, map } from 'rxjs/operators';
import { environment } from '../../environments/environment';

export interface WsMessage {
  type: string;
  payload: any;
}

@Injectable({
  providedIn: 'root'
})
export class WebSocketService implements OnDestroy {
  private socket: WebSocket | null = null;
  private messagesSubject = new Subject<WsMessage>();
  private reconnectInterval = 5000;
  private reconnectTimer: any = null;

  messages$ = this.messagesSubject.asObservable();

  connect(): void {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      return;
    }

    const wsUrl = environment.apiUrl.replace(/^http/, 'ws') + '/ws';
    this.socket = new WebSocket(wsUrl);

    this.socket.onopen = () => {
      console.log('WebSocket connected');
      if (this.reconnectTimer) {
        clearTimeout(this.reconnectTimer);
        this.reconnectTimer = null;
      }
    };

    this.socket.onmessage = (event) => {
      try {
        const message: WsMessage = JSON.parse(event.data);
        this.messagesSubject.next(message);
      } catch (e) {
        console.error('WebSocket message parse error:', e);
      }
    };

    this.socket.onclose = () => {
      console.log('WebSocket disconnected, reconnecting...');
      this.scheduleReconnect();
    };

    this.socket.onerror = (error) => {
      console.error('WebSocket error:', error);
      this.socket?.close();
    };
  }

  private scheduleReconnect(): void {
    if (!this.reconnectTimer) {
      this.reconnectTimer = setTimeout(() => {
        this.reconnectTimer = null;
        this.connect();
      }, this.reconnectInterval);
    }
  }

  disconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
  }

  on(type: string): Observable<any> {
    return this.messages$.pipe(
      filter(msg => msg.type === type),
      map(msg => msg.payload)
    );
  }

  ngOnDestroy(): void {
    this.disconnect();
  }
}
