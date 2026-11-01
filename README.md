# allah-verge

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Release](https://img.shields.io/github/v/release/mateika8814/allah-verge?include_prereleases)](https://github.com/mateika8814/allah-verge/releases)

> **Forked by mateika8814** ([https://github.com/mateika8814](https://github.com/mateika8814))  
> **Based on [Clash Verge Rev](https://github.com/clash-verge-rev/clash-verge-rev) (v2.5.9) and [mihomo](https://github.com/MetaCubeX/mihomo)**

---

## О проекте / Overview

**allah-verge** — это открытый форк Clash Verge Rev со встроенной поддержкой фрагментации пакетов и записей TLS ClientHello для обхода блокировок методом глубокого анализа пакетов (DPI / ТСПУ).

### Что добавлено в форке
1. **Ядро (`core/`)**:
   - Движок фрагментации `component/fragment`, оборачивающий исходящие TCP-соединения TLS-прокси.
   - 4 режима работы: `sni`, `random`, `record_fragment`, `both`.
   - Флаг `record-fragment` для совместимости.
   - Настройка интервалов между пакетами (`interval`) в миллисекундах.
   - RFC-совместимый парсер SNI без внешних библиотек.
   - Логирование предупреждений при передаче неизвестного режима с автоматическим безопасным откатом к `sni`.
2. **Интерфейс (GUI)**:
   - **Редактор ноды**: модальное окно «TLS Fragmentation» в списке прокси для индивидуальной настройки на уровне узла.
   - **Глобальные настройки**: карточка «TLS Fragmentation» в `Settings -> Clash Setting` с переключателем «Apply to all TLS nodes».
   - Валидация диапазонов (`length`, `interval`) и автоматическое скрытие настроек для не-TLS нод.
   - Автоматическая инъекция параметров `fragment:` в runtime-конфиг ядра.

---

## Конфигурация / Configuration

### Пример YAML ноды с `fragment`
```yaml
proxies:
  - name: "Finland-VLESS"
    type: vless
    server: 198.51.100.1
    port: 443
    uuid: 00000000-0000-0000-0000-000000000000
    tls: true
    server-name: example.com
    fragment:
      enabled: true
      mode: sni            # sni | random | record_fragment | both
      length: 10-30        # размер частей в байтах (для random и both)
      interval: 1-5        # пауза между отправками в мс (число или диапазон N-M)
      record-fragment: false # опциональный флаг нарезки записей TLS
```

### Таблица режимов и обработка комбинаций

| Режим (`mode`) | Поведение на проводе | Нарезка TCP | Нарезка записей TLS | `length` |
| :--- | :--- | :---: | :---: | :---: |
| `sni` | Разрезает TCP-сегмент точно по границе SNI (домен обрывается посередине) | Да | Нет | Игнорируется / опускается |
| `random` | Разрезает TCP-сегмент на случайные фрагменты в пределах диапазона `length` | Да | Нет | Обязателен (например, `10-30`) |
| `record_fragment` | Разделяет ClientHello на несколько записей TLS (Content Type 0x16) | Зависит от MTU | Да | Размер блоков записи |
| `both` | Выполняет нарезку на уровне записей TLS **И** передаёт их отдельными TCP-пакетами с задержкой | Да | Да | Размер блоков (например, `10-30`) |

#### Поведение при конфликтах комбинаций
- Если указан флаг `record-fragment: true` совместно с `mode: sni` или `mode: random`, ядро включает нарезку на уровне записей TLS (комбинированный режим, аналогичный `both`).
- Если передан неизвестный или некорректный режим, ядро пишет в лог предупреждение:  
  `[Fragment] unknown fragmentation mode '<value>', falling back to 'sni'`  
  и безопасно переключается на проверенный режим `sni`.

---

## Сборка из исходников / Building

### 1. Сборка ядра (`core/`)
Требуется Go 1.22+.
```bash
cd core
go build -v -trimpath -ldflags "-s -w -buildid=" -tags with_gvisor -o verge-mihomo.exe .
```

### 2. Сборка GUI
Требуются Node.js 18+, pnpm и Rust toolchain.
```bash
# Установка зависимостей
pnpm install

# Запуск в режиме разработки
pnpm dev

# Сборка приложения
pnpm tauri build
```

### 3. Подстановка ядра (Sidecar)
- Для portable-версии: поместите собранный `verge-mihomo.exe` в корень папки приложения рядом с `clash-verge.exe`.
- Для сборки инсталлятора: поместите бинарник в `src-tauri/sidecar/verge-mihomo-x86_64-pc-windows-msvc.exe`.

---

## Загрузка и проверка контрольных сумм / Download & SHA256

Готовая portable-сборка доступна в разделе [Releases](https://github.com/mateika8814/allah-verge/releases).

### Проверка SHA256 в PowerShell
```powershell
Get-FileHash .\Clash-Verge-Rev-Portable-v2.5.9-Fragment-Fork.zip -Algorithm SHA256
```
Сверьте полученный хеш со значением в файле `SHA256SUMS.txt`.

---

## Синхронизация с апстримом / Updating from Upstream

### Обновление GUI
```bash
git fetch upstream
git checkout dev
git pull upstream dev
git checkout feature/tls-fragment
git rebase dev
```

### Обновление ядра по patch
Все изменения ядра хранятся в виде чистого diff-патча [`core/fragment.patch`](./core/fragment.patch):
```bash
cd core
git apply fragment.patch
```

---

## Честные ограничения / Limitations & Transparency

- **Зависимость от провайдера**: Эффективность фрагментации против ТСПУ/DPI напрямую зависит от реализации DPI у конкретного оператора связи и его правил фильтрации. Никаких обещаний «100% гарантированного обхода» не даётся.
- **Физический уровень**: Нарезка пакетов на физическом сетевом интерфейсе (Ethernet) подтверждена дампами Npcap/Wireshark для всех четырёх режимов (`sni`, `random`, `record_fragment`, `both`).
- **Совместимость с Reality**: Режимы с несколькими TLS-записями (`record_fragment`, `both`) являются экспериментальными при работе с серверами VLESS Reality (uTLS), так как некоторые серверные реализации строго проверяют единую TLS-запись в первом пакете. При возникновении ошибок рукопожатия используйте режим `sni`.

---

## Исходный код и лицензия (GPLv3)

- Все релизные бинарники собраны строго из исходного кода данного репозитория на соответствующем теге релиза.
- Исходный код распространяется на условиях лицензии **GNU General Public License v3.0**.
- Подробная информация об авторских правах и апстримах приведена в файле [NOTICE.md](./NOTICE.md).
