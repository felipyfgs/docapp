<?php

use App\Http\Controllers\NfeController;
use Illuminate\Support\Facades\Route;

Route::prefix('v1/nfe')->group(function () {
    Route::post('/distdfe', [NfeController::class, 'distDFe']);
    Route::post('/download', [NfeController::class, 'downloadByKey']);
    Route::post('/consulta-chave', [NfeController::class, 'consultaChave']);
});
