import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router } from '@angular/router';
import { AuthService } from '../../services/auth';
import { filter, take } from 'rxjs/operators';

@Component({
  selector: 'app-cli-auth',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './cli-auth.html'
})
export class CLIAuthComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private authService = inject(AuthService);
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

    // Wait for auth check to complete before reading login state
    this.authService.authCheckComplete$.pipe(
      filter(complete => complete),
      take(1)
    ).subscribe(() => {
      if (!this.authService.isLoggedIn()) {
        // Store port so oauth-callback can restore it after OAuth login
        sessionStorage.setItem('cli_auth_port', this.port);
        this.router.navigate(['/login'], {
          queryParams: { returnUrl: `/cli/auth?port=${this.port}` }
        });
        return;
      }

      // Fetch the CLI token from the backend
      this.authService.getAuthToken().subscribe({
        next: (resp) => {
          this.transferToken(resp.token);
        },
        error: () => {
          this.status = 'error';
          this.errorMessage = 'Failed to retrieve authentication token. Please ensure you are logged in.';
        }
      });
    });
  }

  transferToken(token: string): void {
    // Navigate the browser to the CLI callback URL.
    // Using window.location.href avoids mixed-content blocking
    // (HTTPS→HTTP navigation is allowed; XHR/fetch is not).
    const callbackUrl = `http://127.0.0.1:${this.port}/callback?token=${encodeURIComponent(token)}`;
    this.status = 'success';
    window.location.href = callbackUrl;
  }

  retry(): void {
    window.location.reload();
  }
}
