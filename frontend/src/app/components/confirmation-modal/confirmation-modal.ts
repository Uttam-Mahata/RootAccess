import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ConfirmationModalService, ConfirmationModalData } from '../../services/confirmation-modal.service';
import { Subscription } from 'rxjs';

@Component({
  selector: 'app-confirmation-modal',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './confirmation-modal.html',
  styleUrls: ['./confirmation-modal.scss']
})
export class ConfirmationModalComponent implements OnInit, OnDestroy {
  modalData: ConfirmationModalData | null = null;
  private subscription?: Subscription;

  constructor(private confirmationService: ConfirmationModalService) {}

  ngOnInit(): void {
    this.subscription = this.confirmationService.getModalData().subscribe(data => {
      this.modalData = data;
    });
  }

  ngOnDestroy(): void {
    this.subscription?.unsubscribe();
  }

  onConfirm(): void {
    this.confirmationService.confirm();
  }

  onCancel(): void {
    this.confirmationService.cancel();
  }

  onBackdropClick(): void {
    this.confirmationService.cancel();
  }
}
