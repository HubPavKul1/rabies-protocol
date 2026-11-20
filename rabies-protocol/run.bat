@echo off
chcp 65001 >nul
echo ===================================================
echo   ⚙️  Быстрая сборка и запуск RabiesProtocol
echo ===================================================
echo.

:: Очистка предыдущей сборки (опционально, для чистоты)
if exist build\bin\RabiesProtocol.exe del build\bin\RabiesProtocol.exe

echo [1/2] Сборка приложения (займет ~10-15 секунд)...
wails build

if %errorlevel% neq 0 (
    echo.
    echo ❌ Ошибка сборки! Проверьте код выше.
    pause
    exit /b
)

echo.
echo [2/2] ✅ Сборка успешна! Запуск приложения...
echo.

:: Запуск собранного exe-файла
start "" "build\bin\RabiesProtocol.exe"

echo Приложение запущено. Это окно можно закрыть.