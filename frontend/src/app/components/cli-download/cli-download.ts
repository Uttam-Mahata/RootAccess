import { Component, inject, OnInit, PLATFORM_ID } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';

const DEFAULT_BASE = 'https://ctf.rootaccess.live';

@Component({
  selector: 'app-cli-download',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './cli-download.html'
})
export class CLIDownloadComponent implements OnInit {
  private platformId = inject(PLATFORM_ID);

  copiedValue = '';

  installCommand = `curl -s ${DEFAULT_BASE}/install.sh | bash`;
  apiBaseUrl = DEFAULT_BASE + '/api';

  buildFromSource = `git clone https://github.com/Uttam-Mahata/RootAccess.git
cd RootAccess/cli
go build -o rootaccess main.go
sudo mv rootaccess /usr/local/bin/`;

  crossCompile = `GOOS=linux   GOARCH=amd64 go build -o rootaccess-linux-amd64
GOOS=linux   GOARCH=arm64 go build -o rootaccess-linux-arm64
GOOS=darwin  GOARCH=amd64 go build -o rootaccess-darwin-amd64
GOOS=darwin  GOARCH=arm64 go build -o rootaccess-darwin-arm64
GOOS=windows GOARCH=amd64 go build -o rootaccess-windows-amd64.exe`;

  typicalWorkflow = `# 1. Authenticate
rootaccess login

# 2. Browse challenges
rootaccess challenges

# 3. Read a challenge
rootaccess open 64f1a2b3c4d5e6f7a8b9c0d1

# 4. Work the challenge... then submit your flag
rootaccess submit 64f1a2b3c4d5e6f7a8b9c0d1 'rootaccess{y0ur_fl4g_h3r3}'

# 5. Check your ranking
rootaccess scoreboard

# 6. Confirm your identity at any time
rootaccess whoami

# 7. Log out when done
rootaccess logout`;

  binaries = [
    { name: 'Linux (AMD64)', arch: 'x86_64', file: 'rootaccess-linux-amd64' },
    { name: 'Linux (ARM64)', arch: 'aarch64', file: 'rootaccess-linux-arm64' },
    { name: 'macOS (Intel)', arch: 'x86_64', file: 'rootaccess-darwin-amd64' },
    { name: 'macOS (Apple Silicon)', arch: 'arm64', file: 'rootaccess-darwin-arm64' },
    { name: 'Windows (AMD64)', arch: 'x86_64', file: 'rootaccess-windows-amd64.exe' }
  ];

  private baseUrl = DEFAULT_BASE;

  ngOnInit(): void {
    if (isPlatformBrowser(this.platformId) && typeof window !== 'undefined' && window.location?.origin) {
      this.baseUrl = window.location.origin;
      this.installCommand = `curl -s ${this.baseUrl}/install.sh | bash`;
      this.apiBaseUrl = `${this.baseUrl}/api`;
    } else {
      this.installCommand = `curl -s ${DEFAULT_BASE}/install.sh | bash`;
    }
  }

  copyToClipboard(text: string): void {
    if (!text) return;
    navigator.clipboard.writeText(text).then(() => {
      this.copiedValue = text;
      setTimeout(() => { this.copiedValue = ''; }, 2000);
    });
  }

  binUrl(file: string): string {
    return `${this.baseUrl}/bin/${file}`;
  }
}
