import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';

import { Toaster } from './shared/ui/toaster/toaster';

@Component({
  selector: 'app-root',
  imports: [RouterOutlet, Toaster],
  styleUrl: './app.css',
  templateUrl: './app.html',
})
export class App {}
