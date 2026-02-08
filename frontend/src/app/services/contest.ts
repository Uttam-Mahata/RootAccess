import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, BehaviorSubject, interval } from 'rxjs';
import { tap, switchMap, startWith } from 'rxjs/operators';
import { environment } from '../../environments/environment';

export interface ContestStatus {
  status: 'not_started' | 'running' | 'ended';
  title?: string;
  start_time?: string;
  end_time?: string;
  is_active?: boolean;
}

export interface ContestConfig {
  id: string;
  title: string;
  start_time: string;
  end_time: string;
  is_active: boolean;
}

@Injectable({
  providedIn: 'root'
})
export class ContestService {
  private apiUrl = environment.apiUrl;
  private contestStatusSubject = new BehaviorSubject<ContestStatus | null>(null);
  
  contestStatus$ = this.contestStatusSubject.asObservable();

  constructor(private http: HttpClient) {}

  getContestStatus(): Observable<ContestStatus> {
    return this.http.get<ContestStatus>(`${this.apiUrl}/contest/status`).pipe(
      tap(status => this.contestStatusSubject.next(status))
    );
  }

  startPolling(intervalMs: number = 30000): Observable<ContestStatus> {
    return interval(intervalMs).pipe(
      startWith(0),
      switchMap(() => this.getContestStatus())
    );
  }

  // Admin methods
  getContestConfig(): Observable<any> {
    return this.http.get<any>(`${this.apiUrl}/admin/contest`);
  }

  updateContestConfig(title: string, startTime: string, endTime: string, isActive: boolean): Observable<any> {
    return this.http.put<any>(`${this.apiUrl}/admin/contest`, {
      title,
      start_time: startTime,
      end_time: endTime,
      is_active: isActive
    });
  }

  getCurrentStatus(): ContestStatus | null {
    return this.contestStatusSubject.value;
  }

  getTimeRemaining(): { days: number; hours: number; minutes: number; seconds: number } | null {
    const status = this.contestStatusSubject.value;
    if (!status || !status.end_time) return null;

    const now = new Date().getTime();
    const end = new Date(status.end_time).getTime();
    const diff = end - now;

    if (diff <= 0) return { days: 0, hours: 0, minutes: 0, seconds: 0 };

    return {
      days: Math.floor(diff / (1000 * 60 * 60 * 24)),
      hours: Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60)),
      minutes: Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60)),
      seconds: Math.floor((diff % (1000 * 60)) / 1000)
    };
  }
}
