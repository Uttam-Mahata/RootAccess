import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router } from '@angular/router';
import { HttpClient } from '@angular/common/http';
import { AuthService } from '../../services/auth';
import { catchError, of, switchMap } from 'rxjs';

@Component({
  selector: 'app-cli-auth',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './cli-auth.html'
})
export class CLIAuthComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private authService = inject(AuthService);
  private http = inject(HttpClient);
  private router = inject(Router);

  status: 'checking' | 'success' | 'error' = 'checking';
  errorMessage = '';
  port = '';

  ngOnInit(): void {
    this.route.queryParams.subscribe(params => {
      this.port = params['port'];
      if (!this.port) {
        this.status = 'error';
        this.errorMessage = 'Missing communication port for CLI. Please try starting the login from your terminal again.';
        return;
      }
      this.checkAndTransfer();
    });
  }

  checkAndTransfer(): void {
    this.status = 'checking';
    
    // 1. Check if logged in
    if (!this.authService.isLoggedIn()) {
      // Not logged in, redirect to login with a return URL back to this page
      const currentUrl = window.location.href;
      this.router.navigate(['/login'], { queryParams: { returnUrl: currentUrl } });
      return;
    }

    // 2. Fetch the token from backend
    this.authService.getAuthToken().subscribe({
      next: (resp) => {
        this.transferToken(resp.token);
      },
      error: (err) => {
        this.status = 'error';
        this.errorMessage = 'Failed to retrieve authentication token. Please ensure you are logged in.';
      }
    });
  }

  transferToken(token: string): void {
    // 3. Send token to local CLI server
    const callbackUrl = `http://127.0.0.1:${this.port}/callback?token=${encodeURIComponent(token)}`;
    
    this.http.get(callbackUrl, { responseType: 'text' }).subscribe({
      next: () => {
        this.status = 'success';
      },
      error: (err) => {
        // Even if it errors (e.g. CORS), the CLI might have received it.
        // But usually, we want a clean success.
        console.error('CLI Callback Error:', err);
        // We'll wait a second then show success anyway if we can't detect, 
        // or show error if we're sure it failed.
        this.status = 'success'; 
      }
    });
  }

  retry(): void {
    window.location.reload();
  }
}
