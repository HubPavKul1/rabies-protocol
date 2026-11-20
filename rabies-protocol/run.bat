@echo off
chcp 65001 >nul
echo ===================================================
echo   ⚙️  Сборка и запуск RabiesProtocol (с DevTools)
echo ===================================================
echo.

if exist build\bin\RabiesProtocol.exe del build\bin\RabiesProtocol.exe

echo [1/2] Сборка приложения с включенной панелью разработчика...
wails build -devtools

if %errorlevel% neq 0 (
    echo.
    echo ❌ Ошибка сборки! Проверьте код выше.
    pause
    exit /b
)

echo.
echo [2/2] ✅ Сборка успешна! Запуск приложения...
echo.
echo 💡 Подсказка: Нажмите F12 внутри окна приложения, чтобы открыть консоль.
echo.

start "" "build\bin\RabiesProtocol.exe"