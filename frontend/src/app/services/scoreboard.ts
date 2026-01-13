import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class ScoreboardService {
  private apiUrl = environment.apiUrl;

  constructor(private http: HttpClient) { }

  getScoreboard(): Observable<any[]> {
    return this.http.get<any[]>(`${this.apiUrl}/scoreboard`);
  }
}
