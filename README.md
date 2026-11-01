<h1 align="center">
  <img src="./src-tauri/icons/icon.png" alt="Clash" width="128" />
  <br>
  allah-verge
</h1>

<p align="center">
  <a href="https://www.gnu.org/licenses/gpl-3.0"><img src="https://img.shields.io/badge/License-GPLv3-blue.svg" alt="License: GPL v3" /></a>
  <a href="https://github.com/mateika8814/allah-verge/releases"><img src="https://img.shields.io/github/v/release/mateika8814/allah-verge?include_prereleases" alt="Release" /></a>
</p>

<h3 align="center">
  Forked by <a href="https://github.com/mateika8814">mateika8814</a><br>
  Based on <a href="https://github.com/clash-verge-rev/clash-verge-rev">Clash Verge Rev</a> (v2.5.9, commit <code>ef784e459d31fe9d1946568b6fe47ac3a7e692b7</code>) and <a href="https://github.com/MetaCubeX/mihomo">mihomo</a> (commit <code>88dcbf7f1614a67c3b36b848ee3592dfa92ada36</code>) under GPLv3.
</h3>

---

## Preview

| Dark                             | Light                             |
| -------------------------------- | --------------------------------- |
| ![预览](./docs/preview_dark.png) | ![预览](./docs/preview_light.png) |

---

## TLS ClientHello Fragmentation (DPI / ТСПУ Bypass)

Этот форк добавляет полноценную поддержку нарезки пакетов и TLS-записей ClientHello для обхода систем глубокого анализа пакетов (DPI / ТСПУ).

### Что добавлено
1. **Ядро (`core/`)**:
   - Движок `component/fragment`, оборачивающий исходящие TCP-соединения TLS-нод.
   - 4 режима нарезки: `sni`, `random`, `record_fragment`, `both`.
   - Защита от ложного срабатывания: при первых байтах не-TLS потока или несовпадении заголовка ClientHello (проверка первых 6 байт) данные передаются прозрачно без задержки.
   - Защита от зависания: настраиваемый таймаут буферизации (2 секунды) с автоматическим сбросом накопленных данных без утечек и гонок.
   - Корректная обработка интервалов: независимый расчёт задержки для каждой пары отправляемых сегментов в заданном диапазоне `interval`.
   - Кроссплатформенная поддержка: сборка и работа проверены на Linux, macOS (Darwin) и Windows.
2. **Интерфейс (GUI)**:
   - Модальное окно «TLS Fragmentation» в списке прокси для индивидуальной настройки на уровне узла.
   - Глобальная карточка «TLS Fragmentation» в `Settings -> Clash Setting` с опцией применения ко всем TLS-нодам.
   - Автоматическая инъекция параметров в runtime-конфигурацию ядра.

### Пример YAML ноды
```yaml
proxies:
  - name: "My-TLS-Node"
    type: vless
    server: 198.51.100.1
    port: 443
    uuid: 00000000-0000-0000-0000-000000000000
    tls: true
    server-name: example.com
    fragment:
      enabled: true
      mode: sni              # Режим: sni (по умолчанию) | random | record_fragment | both
      length: 10-30          # Размер фрагментов в байтах (для random и both)
      interval: 1-5          # Пауза между отправками в мс (число или диапазон N-M)
      record-fragment: false # Флаг нарезки записей TLS (для обратной совместимости)
```

### Таблица режимов и поведение

| Режим (`mode`) | Поведение на проводе | Нарезка TCP | Нарезка записей TLS | `length` |
| :--- | :--- | :---: | :---: | :---: |
| **`sni`** *(по умолчанию)* | Разрезает TCP-сегмент ClientHello точно по границе расширения SNI (имя домена обрывается посередине) | Да | Нет | Не требуется / опускается |
| **`random`** | Разрезает ClientHello на случайные сегменты в пределах диапазона `length` | Да | Нет | Обязателен (например, `10-30`) |
| **`record_fragment`** *(экспер.)* | Разделяет ClientHello на несколько валидных TLS-записей (Content Type 0x16) | Зависит от MTU | Да | Размер блоков записи |
| **`both`** *(экспер.)* | Выполняет нарезку на уровне TLS-записей **И** нарезает их на отдельные TCP-пакеты с задержкой | Да | Да | Размер блоков (например, `10-30`) |

#### Поведение при конфликтах комбинаций
- Если указан флаг `record-fragment: true` совместно с `mode: sni` или `mode: random`, флаг `record-fragment` имеет приоритет и активирует разделение на уровне TLS-записей.
- Если режим не указан или передан пустой строкой, по умолчанию используется наиболее стабильный режим **`sni`**.
- Если передано неизвестное значение `mode`, ядро пишет предупреждение в лог (`[Fragment] unknown fragmentation mode '...', falling back to 'sni'`) и безопасно переключается на режим `sni`.
- **Экспериментальный статус**: режимы `record_fragment` и `both` создают множественные TLS-записи и на реальных серверах VLESS Reality могут приводить к сбросу рукопожатия, если сервер требует строгую единую запись в первом пакете. При возникновении проблем используйте стандартный режим `sni`.

---

## Install / Установка

Готовые бинарные сборки форка доступны на странице релизов нашего проекта:  
👉 **[Релизы allah-verge](https://github.com/mateika8814/allah-verge/releases)**

> [!WARNING]
> **Несовместимость со стоковыми версиями**: Стоковый Clash Verge Rev и оригинальное ядро Mihomo не поддерживают блок конфигурации `fragment:`. Для работы фрагментации необходимо использовать сборку из релизов этого форка или собрать ядро из папки `core/`.

---

## Сборка из исходников / Building

### 1. Сборка ядра (`core/`)
Требуется Go 1.22+. Сборка кроссплатформенная (Windows, Linux, macOS):
```bash
cd core
go build -v -trimpath -ldflags "-s -w -buildid=" -tags with_gvisor -o verge-mihomo.exe .
```

### 2. Сборка GUI
Требуются Node.js 18+, pnpm и Rust toolchain:
```bash
pnpm install
pnpm dev
pnpm tauri build
```

### 3. Подстановка ядра (Sidecar)
- Для portable-версии: поместите собранный `verge-mihomo.exe` в корень папки приложения рядом с `clash-verge.exe`.
- Для сборки инсталлятора: поместите бинарник в `src-tauri/sidecar/verge-mihomo-x86_64-pc-windows-msvc.exe`.

---

## Синхронизация с апстримом / Updating from Upstream

```bash
# Добавление апстрима
git remote add upstream https://github.com/clash-verge-rev/clash-verge-rev.git

# Получение обновлений GUI
git fetch upstream
git checkout dev
git pull upstream dev
git checkout feature/tls-fragment
git rebase dev
```

Все изменения ядра зафиксированы в виде чистого патча [`core/fragment.patch`](./core/fragment.patch):
```bash
cd core
git apply fragment.patch
```

---

## Честные ограничения / Limitations & Warranty

- **Зависимость от оператора**: Эффективность фрагментации против систем ТСПУ/DPI зависит от используемого провайдером оборудования и актуальных сигнатур фильтрации. Никаких гарантий «100% обхода всех блокировок» не предоставляется.
- **Физический уровень**: Нарезка пакетов на физическом сетевом интерфейсе (Ethernet) подтверждена дампами Npcap/Wireshark.
- **Отказ от гарантий (GPLv3)**: Программа предоставляется «КАК ЕСТЬ», без каких-либо явных или подразумеваемых гарантий, в том числе гарантий товарной пригодности или применимости для конкретных целей.

---

## Features

- Основан на Rust и Tauri 2.
- Встроенное ядро [mihomo](https://github.com/MetaCubeX/mihomo) с поддержкой фрагментации TLS ClientHello.
- Удобный интерфейс с переключением светлой/тёмной темы и кастомными стилями.
- Управление профилями (Merge, Script), подсветка синтаксиса.
- Системный прокси, TUN-режим (виртуальная сетевая карта).
- Редактирование правил и прокси.

### FAQ

Подробная документация доступна на [странице документации апстрима](https://clash-verge-rev.github.io/faq/windows.html).

---

## Acknowledgement

Clash Verge Rev и allah-verge основаны на разработках следующих проектов:
- [clash-verge-rev/clash-verge-rev](https://github.com/clash-verge-rev/clash-verge-rev): Основной апстрим GUI.
- [MetaCubeX/mihomo](https://github.com/MetaCubeX/mihomo): Ядро прокси и сетевого туннелирования.
- [zzzgydi/clash-verge](https://github.com/zzzgydi/clash-verge): Исходный проект Clash Verge на Tauri.
- [tauri-apps/tauri](https://github.com/tauri-apps/tauri): Фреймворк для настольных приложений.

## Privacy

allah-verge не собирает пользовательские данные. Конфигурации и журналы сохраняются исключительно локально на вашем устройстве.

## License

GNU General Public License v3.0. Подробности см. в файле [LICENSE](./LICENSE) и [NOTICE.md](./NOTICE.md).
