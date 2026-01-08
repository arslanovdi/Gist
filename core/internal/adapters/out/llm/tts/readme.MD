# Gemini-TTS
Есть в [Google AI](https://ai.google.dev/gemini-api/docs/speech-generation?hl=ru#supported-models)
и в [Google Cloud](https://docs.cloud.google.com/text-to-speech/docs/gemini-tts#available_languages)

Функционал несколько отличается.

### Варианты голосового управления VoiceName
Модели TTS поддерживают следующие 30 вариантов голоса в поле voice_name

<div style="display: flex; gap: 20px;">

<div style="flex: 1;">

#### Женские голоса

| Voice Name   | Style                   |
|--------------|-------------------------|
| Achernar     | Мягкий (Soft)           |
| Aoede        | Легкий (Breezy)         |
| Autonoe      | Яркий (Bright)          |
| Callirrhoe   | Спокойный (Easy-going)  |
| Despina      | Плавный (Smooth)        |
| Erinome      | Чистый (Clear)          |
| Gacrux       | Зрелый (Mature)         |
| Kore         | Твердый (Firm)          |
| Laomedeia    | Жизнерадостный (Upbeat) |
| Leda         | Молодой (Youthful)      |
| Pulcherrima  | Прямой (Forward)        |
| Sulafat      | Теплый (Warm)           |
| Vindemiatrix | Нежный (Gentle)         |
| Zephyr       | Яркий (Bright)          |

</div>

<div style="flex: 1;">

#### Мужские голоса

| Voice Name   | Style                              |
|--------------|------------------------------------|
| Achird       | Дружелюбный (Friendly)             |
| Algenib      | Хрипловатый (Gravelly)             |
| Algieba      | Плавный (Smooth)                   |
| Alnilam      | Твердый (Firm)                     |
| Charon       | Информативный (Informative)        |
| Enceladus    | Хрипловатый (Breathy)              |
| Fenrir       | Возбужденный (Excitable)           |
| Iapetus      | Чистый (Clear)                     |
| Orus         | Твердый (Firm)                     |
| Puck         | Жизнерадостный (Upbeat)            |
| Rasalgethi   | Информативный (Informative)        |
| Sadachbia    | Оживлённый (Lively)                |
| Sadaltager   | Знающий специалист (Knowledgeable) |
| Schedar      | Ровный (Even)                      |
| Umbriel      | Спокойный (Easy-going)             |
| Zubenelgenubi| Неформальный (Casual)              |

</div>

</div>

## Google AI
Модели TTS могут только принимать текстовый ввод и генерировать аудиовыход.

В рамках сессии синтеза речи (TTS) существует ограничение на контекстное окно в 32 000 токенов.

### Поддерживаемые модели
- gemini-2.5-flash-preview-tts
- gemini-2.5-pro-preview-tts

Пока [gemini-2.5-flash-preview-tts](https://ai.google.dev/gemini-api/docs/pricing#gemini-2.5-flash-preview-tts) бесплатно и платно.
free tier input quota: 10000 token

### Поддерживаемые языки LanguageCode
Модели TTS автоматически определяют язык ввода. Они поддерживают следующие 24 языка.

Это скорее указание **Акцента**!

| Язык | Код BCP-47 | Язык | Код BCP-47 |
|------|------------|------|------------|
| Арабский (египетский) | ar-EG | Немецкий (Германия) | de-DE |
| Английский (США) | en-US | Испанский (США) | es-US |
| Французский (Франция) | fr-FR | Хинди (Индия) | hi-IN |
| Индонезийский (Индонезия) | id-ID | Итальянский (Италия) | it-IT |
| Японский (Япония) | ja-JP | Корейский (Корея) | ko-KR |
| Португальский (Бразилия) | pt-BR | Русский (Россия) | ru-RU |
| Голландский (Нидерланды) | nl-NL | Польский (Польша) | pl-PL |
| Тайский (Таиланд) | th-TH | Турецкий (Турция) | tr-TR |
| Вьетнамский (Вьетнам) | vi-VN | Румынский (Румыния) | ro-RO |
| Украинский (Украина) | uk-UA | Бенгальский (Бангладеш) | bn-BD |
| Английский (Индия) | en-IN & hi-IN bundle | Маратхи (Индия) | mr-IN |
| Тамильский (Индия) | ta-IN | Телугу (Индия) | te-IN |

### Поддержка многоголосого вывода

## Google Cloud

В этих моделях есть возможность выбрать выходной формат.
Больше поддерживаемых языков, есть Preview языки.

Максимальная длительность выходного аудиосигнала 655 секунд. Если больше, то обрезается.

### Поддерживаемые модели
- gemini-2.5-flash-tts
- gemini-2.5-flash-lite-preview-tts
- gemini-2.5-pro-tts

### Поддерживаемые выходные аудиоформаты
- Унарный: LINEAR16 (по умолчанию), ALAW, MULAW, MP3, OGG_OPUS, PCM
- Потоковый: PCM (по умолчанию), ALAW, MULAW, OGG_OPUS

### Поддерживаемые языки LanguageCode

Модели TTS автоматически определяют язык ввода. Они поддерживают следующие 24 языка.

Это скорее указание Акцента!

| Language | BCP-47 Code | Status   |
|----------|-------------|----------|
| Arabic (Egypt) | ar-EG | GA |
| Bangla (Bangladesh) | bn-BD | GA |
| Dutch (Netherlands) | nl-NL | GA |
| English (India) | en-IN | GA |
| English (United States) | en-US | GA |
| French (France) | fr-FR | GA |
| German (Germany) | de-DE | GA |
| Hindi (India) | hi-IN | GA |
| Indonesian (Indonesia) | id-ID | GA |
| Italian (Italy) | it-IT | GA |
| Japanese (Japan) | ja-JP | GA |
| Korean (South Korea) | ko-KR | GA |
| Marathi (India) | mr-IN | GA |
| Polish (Poland) | pl-PL | GA |
| Portuguese (Brazil) | pt-BR | GA |
| Romanian (Romania) | ro-RO | GA |
| Russian (Russia) | ru-RU | GA |
| Spanish (Spain) | es-ES | GA |
| Tamil (India) | ta-IN | GA |
| Telugu (India) | te-IN | GA |
| Thai (Thailand) | th-TH | GA |
| Turkish (Turkey) | tr-TR | GA |
| Ukrainian (Ukraine) | uk-UA | GA |
| Vietnamese (Vietnam) | vi-VN | GA |
| Afrikaans (South Africa) | af-ZA | Preview |
| Albanian (Albania) | sq-AL | Preview |
| Amharic (Ethiopia) | am-ET | Preview |
| Arabic (World) | ar-001 | Preview |
| Armenian (Armenia) | hy-AM | Preview |
| Azerbaijani (Azerbaijan) | az-AZ | Preview |
| Basque (Spain) | eu-ES | Preview |
| Belarusian (Belarus) | be-BY | Preview |
| Bulgarian (Bulgaria) | bg-BG | Preview |
| Burmese (Myanmar) | my-MM | Preview |
| Catalan (Spain) | ca-ES | Preview |
| Cebuano (Philippines) | ceb-PH | Preview |
| Chinese, Mandarin (China) | cmn-CN | Preview |
| Chinese, Mandarin (Taiwan) | cmn-tw | Preview |
| Croatian (Croatia) | hr-HR | Preview |
| Czech (Czech Republic) | cs-CZ | Preview |
| Danish (Denmark) | da-DK | Preview |
| English (Australia) | en-AU | Preview |
| English (United Kingdom) | en-GB | Preview |
| Estonian (Estonia) | et-EE | Preview |
| Filipino (Philippines) | fil-PH | Preview |
| Finnish (Finland) | fi-FI | Preview |
| French (Canada) | fr-CA | Preview |
| Galician (Spain) | gl-ES | Preview |
| Georgian (Georgia) | ka-GE | Preview |
| Greek (Greece) | el-GR | Preview |
| Gujarati (India) | gu-IN | Preview |
| Haitian Creole (Haiti) | ht-HT | Preview |
| Hebrew (Israel) | he-IL | Preview |
| Hungarian (Hungary) | hu-HU | Preview |
| Icelandic (Iceland) | is-IS | Preview |
| Javanese (Java) | jv-JV | Preview |
| Kannada (India) | kn-IN | Preview |
| Konkani (India) | kok-IN | Preview |
| Lao (Laos) | lo-LA | Preview |
| Latin (Vatican City) | la-VA | Preview |
| Latvian (Latvia) | lv-LV | Preview |
| Lithuanian (Lithuania) | lt-LT | Preview |
| Luxembourgish (Luxembourg) | lb-LU | Preview |
| Macedonian (North Macedonia) | mk-MK | Preview |
| Maithili (India) | mai-IN | Preview |
| Malagasy (Madagascar) | mg-MG | Preview |
| Malay (Malaysia) | ms-MY | Preview |
| Malayalam (India) | ml-IN | Preview |
| Mongolian (Mongolia) | mn-MN | Preview |
| Nepali (Nepal) | ne-NP | Preview |
| Norwegian, Bokmål (Norway) | nb-NO | Preview |
| Norwegian, Nynorsk (Norway) | nn-NO | Preview |
| Odia (India) | or-IN | Preview |
| Pashto (Afghanistan) | ps-AF | Preview |
| Persian (Iran) | fa-IR | Preview |
| Portuguese (Portugal) | pt-PT | Preview |
| Punjabi (India) | pa-IN | Preview |
| Serbian (Serbia) | sr-RS | Preview |
| Sindhi (India) | sd-IN | Preview |
| Sinhala (Sri Lanka) | si-LK | Preview |
| Slovak (Slovakia) | sk-SK | Preview |
| Slovenian (Slovenia) | sl-SI | Preview |
| Spanish (Latin America) | es-419 | Preview |
| Spanish (Mexico) | es-MX | Preview |
| Swahili (Kenya) | sw-KE | Preview |
| Swedish (Sweden) | sv-SE | Preview |
| Urdu (Pakistan) | ur-PK | Preview |
