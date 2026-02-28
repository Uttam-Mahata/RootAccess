import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';

@Component({
  selector: 'app-legal',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './legal.html',
  styleUrls: ['./legal.scss']
})
export class LegalComponent {}

