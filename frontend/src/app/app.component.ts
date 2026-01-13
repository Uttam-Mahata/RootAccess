import { Component, inject } from '@angular/core';
import { RouterOutlet, RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';
import { AuthService } from './services/auth';
import { ThemeService } from './services/theme';
import 'zone.js';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, RouterModule, CommonModule],
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent {
  public authService = inject(AuthService);
  public themeService = inject(ThemeService);

  isLoggedIn(): boolean {
    return this.authService.isLoggedIn();
  }

  logout(): void {
    this.authService.logout();
    window.location.reload();
  }

  toggleTheme(): void {
    console.log('Toggle theme clicked', this.themeService.isDarkMode());
    this.themeService.toggleTheme();
    console.log('After toggle', this.themeService.isDarkMode());
  }
}
