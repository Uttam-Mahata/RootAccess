import { Component, Input, forwardRef, ElementRef, ViewChild, AfterViewInit, OnDestroy, OnChanges, SimpleChanges } from '@angular/core';
import { ControlValueAccessor, NG_VALUE_ACCESSOR } from '@angular/forms';
import flatpickr from 'flatpickr';
import { Instance } from 'flatpickr/dist/types/instance';

@Component({
  selector: 'app-datetime-picker',
  standalone: true,
  template: `
    <div class="relative">
      <input #pickerInput
             type="text"
             [placeholder]="placeholder"
             readonly
             class="w-full px-4 py-2 bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded-lg text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-red-600 focus:border-transparent transition-all duration-200 cursor-pointer pr-20"
      />
      <div class="absolute right-2 top-1/2 -translate-y-1/2 flex items-center gap-1">
        <span class="text-xs text-slate-400 dark:text-slate-500 select-none">{{ timezoneLabel }}</span>
        <svg class="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/>
        </svg>
      </div>
    </div>
  `,
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => DatetimePickerComponent),
      multi: true
    }
  ]
})
export class DatetimePickerComponent implements ControlValueAccessor, AfterViewInit, OnDestroy, OnChanges {
  @ViewChild('pickerInput') inputRef!: ElementRef<HTMLInputElement>;
  @Input() placeholder = 'Select date & time';
  @Input() min?: string;
  @Input() max?: string;

  private fp: Instance | null = null;
  private onChange: (value: string) => void = () => {};
  private onTouched: () => void = () => {};
  private currentValue = '';

  timezoneLabel: string;

  constructor() {
    const offset = new Date().getTimezoneOffset();
    const sign = offset <= 0 ? '+' : '-';
    const hours = Math.floor(Math.abs(offset) / 60);
    const mins = Math.abs(offset) % 60;
    this.timezoneLabel = `UTC${sign}${hours}${mins ? ':' + String(mins).padStart(2, '0') : ''}`;
  }

  ngAfterViewInit(): void {
    this.initFlatpickr();
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (this.fp && (changes['min'] || changes['max'])) {
      this.updateMinMax();
    }
  }

  ngOnDestroy(): void {
    this.fp?.destroy();
  }

  private initFlatpickr(): void {
    const isDark = document.documentElement.classList.contains('dark');

    this.fp = flatpickr(this.inputRef.nativeElement, {
      enableTime: true,
      dateFormat: 'Y-m-d H:i',
      altInput: true,
      altFormat: 'M j, Y h:i K',
      time_24hr: false,
      minuteIncrement: 1,
      defaultDate: this.currentValue ? new Date(this.currentValue) : undefined,
      minDate: this.min ? new Date(this.min) : undefined,
      maxDate: this.max ? new Date(this.max) : undefined,
      onChange: (selectedDates) => {
        if (selectedDates.length > 0) {
          const iso = selectedDates[0].toISOString();
          this.currentValue = iso;
          this.onChange(iso);
        }
      },
      onClose: () => {
        this.onTouched();
      }
    }) as Instance;

    // Apply dark mode styling to alt input
    if (this.fp.altInput) {
      this.fp.altInput.className = this.inputRef.nativeElement.className;
    }
  }

  private updateMinMax(): void {
    if (!this.fp) return;
    if (this.min) {
      this.fp.set('minDate', new Date(this.min));
    }
    if (this.max) {
      this.fp.set('maxDate', new Date(this.max));
    }
  }

  writeValue(value: string): void {
    this.currentValue = value || '';
    if (this.fp && value) {
      this.fp.setDate(new Date(value), false);
    } else if (this.fp && !value) {
      this.fp.clear(false);
    }
  }

  registerOnChange(fn: (value: string) => void): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: () => void): void {
    this.onTouched = fn;
  }

  setDisabledState(isDisabled: boolean): void {
    if (this.fp) {
      if (isDisabled) {
        this.fp.altInput?.setAttribute('disabled', 'true');
      } else {
        this.fp.altInput?.removeAttribute('disabled');
      }
    }
  }
}
