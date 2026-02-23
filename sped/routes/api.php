<?php

use App\Http\Controllers\NfeController;
use App\Http\Controllers\SefazController;
use Illuminate\Support\Facades\Route;

Route::prefix('v1/nfe')->group(function () {
    Route::post('/distdfe', [NfeController::class, 'distDFe']);
    Route::post('/download', [NfeController::class, 'downloadByKey']);
    Route::post('/consulta-chave', [NfeController::class, 'consultaChave']);
});

Route::prefix('v1/sefaz')->group(function () {
    Route::post('/distdfe', [SefazController::class, 'distDFe']);
    Route::post('/download', [SefazController::class, 'download']);
    Route::post('/consulta', [SefazController::class, 'consulta']);
    Route::post('/manifesta', [SefazController::class, 'manifesta']);
    Route::post('/danfe', [SefazController::class, 'danfe']);
});
