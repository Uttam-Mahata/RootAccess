import { Injectable, OnDestroy, inject } from '@angular/core';
import { Subject, BehaviorSubject, Observable } from 'rxjs';
import { filter, map } from 'rxjs/operators';
import { environment } from '../../environments/environment';
import { AuthService } from './auth';

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
  private connectedSubject = new BehaviorSubject<boolean>(false);
  private reconnectInterval = 5000;
  private reconnectTimer: any = null;

  private authService = inject(AuthService);

  messages$ = this.messagesSubject.asObservable();
  connected$ = this.connectedSubject.asObservable();

  connect(): void {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      return;
    }

    // Don't connect if not logged in
    if (!this.authService.isLoggedIn()) {
      return;
    }

    // Get token from cookie
    const getCookie = (name: string) => {
      const value = `; ${document.cookie}`;
      const parts = value.split(`; ${name}=`);
      if (parts.length === 2) return parts.pop()?.split(';').shift();
      return null;
    };
    const token = getCookie('auth_token');

    let wsUrl = environment.wsUrl;
    // If we have a token, append it to the URL for AWS Lambda WebSocket auth
    if (token) {
      // Check if URL already has query params
      const separator = wsUrl.includes('?') ? '&' : '?';
      wsUrl += `${separator}token=${token}`;
    }

    try {
      this.socket = new WebSocket(wsUrl);

      this.socket.onopen = () => {
        console.log('WebSocket connected');
        this.connectedSubject.next(true);
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
          // Suppress noise
        }
      };

      this.socket.onclose = (event) => {
        if (this.connectedSubject.value) {
          console.log('WebSocket disconnected, reconnecting...');
        }
        this.connectedSubject.next(false);
        // Only retry if it wasn't a clean close and we are logged in
        if (this.authService.isLoggedIn()) {
          this.scheduleReconnect();
        }
      };

      this.socket.onerror = (error) => {
        // Suppress error noise in console as it's expected on some environments
        this.connectedSubject.next(false);
        this.socket?.close();
      };
    } catch (err) {
      this.connectedSubject.next(false);
    }
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
