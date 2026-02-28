import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';

@Component({
  selector: 'app-cli-download',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './cli-download.html'
})
export class CLIDownloadComponent {
  commandCopied = false;

  binaries = [
    { name: 'Linux (AMD64)', arch: 'x86_64', file: 'rootaccess-linux-amd64' },
    { name: 'Linux (ARM64)', arch: 'aarch64', file: 'rootaccess-linux-arm64' },
    { name: 'macOS (Intel)', arch: 'x86_64', file: 'rootaccess-darwin-amd64' },
    { name: 'macOS (Apple Silicon)', arch: 'arm64', file: 'rootaccess-darwin-arm64' },
    { name: 'Windows (AMD64)', arch: 'x86_64', file: 'rootaccess-windows-amd64.exe' }
  ];

  copyInstallCommand(): void {
    const cmd = 'curl -s https://ctf.rootaccess.live/install.sh | bash';
    navigator.clipboard.writeText(cmd).then(() => {
      this.commandCopied = true;
      setTimeout(() => this.commandCopied = false, 2000);
    });
  }
}
