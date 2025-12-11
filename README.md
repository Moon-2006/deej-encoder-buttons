# Deej with Rotary Encoder Button Support

This is a fork of the [original Deej](https://github.com/omriharel/deej) with additional functionality for **rotary encoder buttons**.

## 📝 About this Fork

This project adds native support for rotary encoder buttons, allowing you to control media functions (play/pause, next, previous, mute) directly from your Deej device.

**⚠️ Important Notice:**
- I am not a professional developer
- This version was created with AI assistance
- Tested on:
  - **Windows 11** with **Arduino Nano v3.0 (atmega328p/168p)**
  - **Linux (Fedora)** with **Arduino Nano v3.0**
- If you encounter bugs or issues, feel free to fork and adapt it to your needs

## 🖥️ Platform Support

- ✅ **Windows** - Full support with native Windows API
- ✅ **Linux** - Full support using playerctl (MPRIS) and pactl (PulseAudio)
  - **Works seamlessly on both X11 and Wayland** - no permission prompts required

## ✨ New Features

- **Button configuration via YAML** - Define button actions without reprogramming the Arduino
- **Media commands** - Play/pause, next track, previous track, master mute
- **Easy customization** - Change functions by editing only the `config.yaml`
- **Cross-platform** - Works on both Windows and Linux

## 🔧 Configuration

### 1. Arduino Code

An example file is included in the project: `arduino/Arduino_Nano_4_encodeers+buttons.ino`

**Important:** Buttons must send commands in this format:
```
BTN:0  // Encoder 1 button
BTN:1  // Encoder 2 button
BTN:2  // Encoder 3 button
BTN:3  // Encoder 4 button
```

### 2. Serial Port Configuration

**⚠️ IMPORTANT:** Don't forget to configure the correct serial port in your `config.yaml`!

**Windows:**
```yaml
com_port: COM4  # Change to your Arduino's COM port (check Device Manager)
```

**Linux:**
```yaml
com_port: /dev/ttyUSB0  # or /dev/ttyACM0 (check with: ls /dev/tty*)
```

To find your Arduino port on Linux:
```bash
ls /dev/tty* | grep -E "(USB|ACM)"
```

### 3. config.yaml Configuration

Change the media controls according to your setup:

```yaml
button_mapping:
  0: mute_master        # Mute/unmute system
  1: media_play_pause   # Play/Pause
  2: media_previous     # Previous track
  3: media_next         # Next track

com_port: COM4  # Windows example (change accordingly)
# com_port: /dev/ttyUSB0  # Linux example (uncomment and change if using Linux)

baud_rate: 9600
```

### 4. Application Configuration (Linux)

On Linux, you can configure sliders to control any application by specifying the audio process name in `slider_mapping`.

#### 🔍 How to Find Application Names

**Step 1:** Open the application you want to control (e.g., a game, music player, browser)

**Step 2:** Make sure the app is playing audio

**Step 3:** Run this command in terminal:
```bash
pactl list sink-inputs | grep -E "(application.name|application.process.binary)"
```

**Example output:**
```
application.name = "Spotify"
application.process.binary = "spotify"

application.name = "Firefox"
application.process.binary = "firefox"

application.name = "CS2"
application.process.binary = "cs2"
```

**Step 4:** Use the value from `application.process.binary` in your config

#### 📋 Common Application Names

**Native Linux Apps:**
```yaml
slider_mapping:
  0: master
  1: spotify          # Spotify
  2: firefox          # Firefox browser
  3: discord          # Discord
```

**Steam/Proton Games (Windows games on Linux):**
```yaml
slider_mapping:
  0: master
  1: spotify
  2: wine64-preloader  # All Proton games (64-bit)
  3: wine-preloader    # Old Proton games (32-bit)
```

**Native Linux Games:**
```yaml
slider_mapping:
  0: master
  1: spotify
  2: dota2            # Dota 2
  3: csgo             # CS:GO
```

#### 💡 Pro Tips

- **Steam/Proton games** all use `wine64-preloader` (or `wine-preloader` for 32-bit games)
- **One slider controls all Proton games** at once (they share the same audio process)
- **Native Linux games** each have their own process name
- **Flatpak apps** usually keep their normal name (e.g., `discord`, `spotify`)

#### 🎮 Special Case: deej.current

The `deej.current` feature attempts to automatically detect and control the currently focused window. However, for **most reliable control of Steam/Proton games**, we recommend using `wine64-preloader` directly:

```yaml
slider_mapping:
  2: wine64-preloader  # Recommended for Steam/Proton games
  # 2: deej.current    # Alternative (may not work with all games)
```

## 📦 Installation

### Windows
1. Download `deej.exe` and `config.yaml` from [Releases](https://github.com/guizmo-silva/deej-encoder-buttons/releases)
2. Place them in the same directory
3. Edit `config.yaml` with your COM port and button mappings
4. Run `deej.exe`

### Linux
1. Download `deej-linux` and `config_linux.yaml` from [Releases](https://github.com/guizmo-silva/deej-encoder-buttons/releases)
2. Make the binary executable: `chmod +x deej-linux`
3. Install playerctl: `sudo apt install playerctl` (Debian/Ubuntu) or `sudo dnf install playerctl` (Fedora)
4. Edit `config.yaml` with your serial port (usually `/dev/ttyUSB0` or `/dev/ttyACM0`)
5. Run `./deej-linux`



## ⚠️ IMPORTANT - Arduino Not Recognized Issues

If Deej doesn't recognize your Arduino, the problem might be with the **CH340 driver**:

1. [Download and install the CH340 driver](https://learn.sparkfun.com/tutorials/how-to-install-ch340-drivers/all)
2. If already installed and still not working: **completely uninstall and reinstall**
3. Restart your computer after installation

## 💬 Questions about assembly?

For questions about building the Deej hardware, join the [official Deej Discord](https://discord.com/invite/nf88NJu)

#### Special thanks to _TheGiantSwede_ from the Deej Discord who helped me with the Arduino code with encoders

## 📄 License

This project maintains the MIT license from the original Deej.

---

# 🇧🇷 Deej com Suporte a Botões para Rotary Encoders

Este é um fork do [Deej original](https://github.com/omriharel/deej) com funcionalidade adicional para **botões de rotary encoders**.

## 📝 Sobre este Fork

Este projeto adiciona suporte nativo aos botões dos rotary encoders, permitindo que você controle funções de mídia (play/pause, next, previous, mute) diretamente pelo seu dispositivo Deej.

**⚠️ Aviso Importante:**
- Não sou desenvolvedor profissional
- Esta versão foi criada com ajuda de IA
- Testada em:
  - **Windows 11** com **Arduino Nano v3.0 (atmega328p/168p)**
  - **Linux (Fedora)** com **Arduino Nano v3.0**
- Se encontrar bugs ou problemas, sinta-se à vontade para fazer um fork e adaptar às suas necessidades

## 🖥️ Suporte a Plataformas

- ✅ **Windows** - Suporte completo com API nativa do Windows
- ✅ **Linux** - Suporte completo usando playerctl (MPRIS) e pactl (PulseAudio)
  - **Funciona perfeitamente tanto no X11 quanto no Wayland** - sem prompts de permissão

## ✨ Novas Funcionalidades

- **Configuração de botões via YAML** - Defina as ações dos botões sem reprogramar o Arduino
- **Comandos de mídia** - Play/pause, próxima música, música anterior, mute master
- **Fácil personalização** - Mude as funções editando apenas o `config.yaml`
- **Multiplataforma** - Funciona tanto no Windows quanto no Linux

## 🔧 Configuração

### 1. Código do Arduino

Um arquivo de exemplo está incluído no projeto: `arduino/Arduino_Nano_4_encodeers+buttons.ino`

**Importante:** Os botões devem enviar comandos no formato:
```
BTN:0  // Botão do encoder 1
BTN:1  // Botão do encoder 2
BTN:2  // Botão do encoder 3
BTN:3  // Botão do encoder 4
```

### 2. Configuração da Porta Serial

**⚠️ IMPORTANTE:** Não esqueça de configurar a porta serial correta no seu `config.yaml`!

**Windows:**
```yaml
com_port: COM4  # Mude para a porta COM do seu Arduino (verifique no Gerenciador de Dispositivos)
```

**Linux:**
```yaml
com_port: /dev/ttyUSB0  # ou /dev/ttyACM0 (verifique com: ls /dev/tty*)
```

Para encontrar a porta do Arduino no Linux:
```bash
ls /dev/tty* | grep -E "(USB|ACM)"
```

### 3. Configuração do config.yaml

Mude os controles de mídia de acordo com a necessidade do seu projeto:

```yaml
button_mapping:
  0: mute_master        # Muta/desmuta o sistema
  1: media_play_pause   # Play/Pause
  2: media_previous     # Música anterior
  3: media_next         # Próxima música

com_port: COM4  # Exemplo Windows (mude conforme necessário)
# com_port: /dev/ttyUSB0  # Exemplo Linux (descomente e mude se estiver usando Linux)

baud_rate: 9600
```

### 4. Configuração de Aplicações (Linux)

No Linux, você pode configurar os sliders para controlar qualquer aplicação especificando o nome do processo de áudio em `slider_mapping`.

#### 🔍 Como Encontrar o Nome das Aplicações

**Passo 1:** Abra a aplicação que você quer controlar (ex: um jogo, player de música, navegador)

**Passo 2:** Certifique-se que o app está tocando áudio

**Passo 3:** Execute este comando no terminal:
```bash
pactl list sink-inputs | grep -E "(application.name|application.process.binary)"
```

**Exemplo de saída:**
```
application.name = "Spotify"
application.process.binary = "spotify"

application.name = "Firefox"
application.process.binary = "firefox"

application.name = "CS2"
application.process.binary = "cs2"
```

**Passo 4:** Use o valor de `application.process.binary` no seu config

#### 📋 Nomes Comuns de Aplicações

**Apps Nativos do Linux:**
```yaml
slider_mapping:
  0: master
  1: spotify          # Spotify
  2: firefox          # Navegador Firefox
  3: discord          # Discord
```

**Jogos Steam/Proton (jogos Windows no Linux):**
```yaml
slider_mapping:
  0: master
  1: spotify
  2: wine64-preloader  # Todos os jogos Proton (64-bit)
  3: wine-preloader    # Jogos Proton antigos (32-bit)
```

**Jogos Nativos do Linux:**
```yaml
slider_mapping:
  0: master
  1: spotify
  2: dota2            # Dota 2
  3: csgo             # CS:GO
```

#### 💡 Dicas Importantes

- **Jogos Steam/Proton** todos usam `wine64-preloader` (ou `wine-preloader` para jogos 32-bit)
- **Um slider controla todos os jogos Proton** de uma vez (eles compartilham o mesmo processo de áudio)
- **Jogos nativos do Linux** cada um tem seu próprio nome de processo
- **Apps Flatpak** geralmente mantêm o nome normal (ex: `discord`, `spotify`)

#### 🎮 Caso Especial: deej.current

A funcionalidade `deej.current` tenta detectar e controlar automaticamente a janela em foco. Porém, para **controle mais confiável de jogos Steam/Proton**, recomendamos usar `wine64-preloader` diretamente:

```yaml
slider_mapping:
  2: wine64-preloader  # Recomendado para jogos Steam/Proton
  # 2: deej.current    # Alternativa (pode não funcionar com todos os jogos)
```

## 📦 Instalação

### Windows
1. Baixe `deej.exe` e `config.yaml` da página de [Releases](https://github.com/guizmo-silva/deej-encoder-buttons/releases)
2. Coloque-os no mesmo diretório
3. Edite `config.yaml` com sua porta COM e mapeamento de botões
4. Execute `deej.exe`

### Linux
1. Baixe `deej-linux` e `config_linux.yaml` da página de [Releases](https://github.com/guizmo-silva/deej-encoder-buttons/releases)
2. Torne o binário executável: `chmod +x deej-linux`
3. Instale o playerctl: `sudo apt install playerctl` (Debian/Ubuntu) ou `sudo dnf install playerctl` (Fedora)
4. Edite `config.yaml` com sua porta serial (geralmente `/dev/ttyUSB0` ou `/dev/ttyACM0`)
5. Execute `./deej-linux`


## ⚠️ IMPORTANTE - Problemas com Arduino não reconhecido

Se o Deej não reconhecer seu Arduino, o problema pode ser com o **driver CH340**:

1. [Baixe e instale o driver CH340](https://learn.sparkfun.com/tutorials/how-to-install-ch340-drivers/all)
2. Se já tiver instalado e ainda assim não funcionar: **desinstale completamente e reinstale**
3. Reinicie o computador após a instalação

## 💬 Dúvidas sobre montagem?

Para dúvidas sobre como montar o hardware do Deej, entre no [Discord oficial do Deej](https://discord.com/invite/nf88NJu)

#### Agradecimento especial ao user _TheGiantSwede_ do Discord do Deej pela ajuda com o código de rotary encoders no Arduino

## 📄 Licença

Este projeto mantém a licença MIT do Deej original.