import { Component, OnInit } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { AuthService } from '../../services/auth';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-oauth-callback',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './oauth-callback.html',
  styleUrls: ['./oauth-callback.scss']
})
export class OAuthCallbackComponent implements OnInit {
  success = false;
  error = '';
  username = '';

  constructor(
    private authService: AuthService,
    private router: Router,
    private route: ActivatedRoute
  ) {}

  ngOnInit(): void {
    // Get query params
    this.route.queryParams.subscribe(params => {
      this.success = params['success'] === 'true';
      this.error = params['error'] || '';
      this.username = params['username'] || '';

      if (this.success) {
        // Check auth status and redirect
        this.authService.checkAuthStatus();

        // If the user came from the CLI auth flow, send them back there
        const cliPort = sessionStorage.getItem('cli_auth_port');
        if (cliPort) {
          sessionStorage.removeItem('cli_auth_port');
          setTimeout(() => {
            this.router.navigate(['/cli/auth'], { queryParams: { port: cliPort } });
          }, 500);
        } else {
          // Wait a moment for auth status to update, then redirect
          setTimeout(() => {
            this.router.navigate(['/challenges']);
          }, 1500);
        }
      } else if (this.error) {
        // Show error for a few seconds, then redirect to login
        setTimeout(() => {
          this.router.navigate(['/login']);
        }, 3000);
      }
    });
  }
}
