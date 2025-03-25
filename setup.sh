#!/bin/bash

set -e

# Variables
GITHUB_REPO="shaharia-lab/echoy"
INSTALL_DIR="$HOME/.echoy"
BIN_DIR="$HOME/.local/bin"
LATEST_VERSION=$(curl -s "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | grep -o '"tag_name": ".*"' | sed 's/"tag_name": "//;s/"//')

# Create message functions with colored output
info() {
  echo -e "\033[0;34m[INFO]\033[0m $1"
}

success() {
  echo -e "\033[0;32m[SUCCESS]\033[0m $1"
}

error() {
  echo -e "\033[0;31m[ERROR]\033[0m $1" >&2
}

warn() {
  echo -e "\033[0;33m[WARNING]\033[0m $1"
}

# Detect OS and architecture
detect_platform() {
  OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
  ARCH="$(uname -m)"

  case "$OS" in
    linux) OS="linux" ;;
    darwin) OS="darwin" ;;
    msys*|mingw*|cygwin*|win*)
      OS="windows"
      if [ -n "$MSYSTEM" ] || [ -n "$CYGWIN" ]; then
        warn "Running in $MSYSTEM/Cygwin environment"
      fi
      ;;
    *)
      error "Unsupported operating system: $OS"
      exit 1
      ;;
  esac

  case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *)
      error "Unsupported architecture: $ARCH"
      exit 1
      ;;
  esac

  info "Detected platform: $OS $ARCH"
}

# Create necessary directories
setup_directories() {
  info "Creating required directories..."
  mkdir -p "$INSTALL_DIR"
  mkdir -p "$BIN_DIR"
}

# Download the binary
download_binary() {
  if [ -z "$LATEST_VERSION" ]; then
    error "Could not determine latest version"
    exit 1
  fi

  # Remove 'v' prefix if present in version
  VERSION_NUM=${LATEST_VERSION#v}

  info "Installing Echoy $LATEST_VERSION for $OS $ARCH..."

  BINARY_NAME="echoy"
  if [ "$OS" = "windows" ]; then
    BINARY_NAME="echoy.exe"
  fi

  DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/$LATEST_VERSION/echoy_${VERSION_NUM}_${OS}_${ARCH}.tar.gz"
  TMP_DIR=$(mktemp -d)

  info "Downloading from: $DOWNLOAD_URL"

  # Download and extract
  if ! curl -L "$DOWNLOAD_URL" -o "$TMP_DIR/echoy.tar.gz"; then
    error "Failed to download binary"
    exit 1
  fi

  tar -xzf "$TMP_DIR/echoy.tar.gz" -C "$TMP_DIR"

  # Move the binary to the install directory
  # First, find the binary in the extracted files
  EXTRACTED_BINARY=$(find "$TMP_DIR" -name "$BINARY_NAME" -type f | head -n 1)

  if [ -z "$EXTRACTED_BINARY" ]; then
    # If we can't find the binary by name, use the first executable file
    EXTRACTED_BINARY=$(find "$TMP_DIR" -type f -perm /111 | head -n 1)
  fi

  if [ -z "$EXTRACTED_BINARY" ]; then
    error "Could not find the executable in the downloaded archive"
    ls -la "$TMP_DIR"
    exit 1
  fi

  mv "$EXTRACTED_BINARY" "$INSTALL_DIR/$BINARY_NAME"
  chmod +x "$INSTALL_DIR/$BINARY_NAME"

  # Clean up
  rm -rf "$TMP_DIR"
}

# Create symlink
create_symlink() {
  BINARY_NAME="echoy"
  if [ "$OS" = "windows" ]; then
    BINARY_NAME="echoy.exe"
  fi

  # Create symlink in bin directory
  if [ -f "$BIN_DIR/$BINARY_NAME" ]; then
    rm "$BIN_DIR/$BINARY_NAME"
  fi

  ln -s "$INSTALL_DIR/$BINARY_NAME" "$BIN_DIR/$BINARY_NAME"

  info "Created symlink to $BIN_DIR/$BINARY_NAME"
}

# Check if PATH includes the bin directory
update_path() {
  if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
    warn "The directory $BIN_DIR is not in your PATH"

    SHELL_CONFIG=""
    if [ -n "$BASH_VERSION" ]; then
      if [ -f "$HOME/.bashrc" ]; then
        SHELL_CONFIG="$HOME/.bashrc"
      elif [ -f "$HOME/.bash_profile" ]; then
        SHELL_CONFIG="$HOME/.bash_profile"
      fi
    elif [ -n "$ZSH_VERSION" ]; then
      SHELL_CONFIG="$HOME/.zshrc"
    fi

    if [ -n "$SHELL_CONFIG" ]; then
      echo "export PATH=\"\$PATH:$BIN_DIR\"" >> "$SHELL_CONFIG"
      info "Added $BIN_DIR to your PATH in $SHELL_CONFIG"
      info "Please restart your terminal or run 'source $SHELL_CONFIG' to apply changes"
    else
      warn "Could not detect shell configuration file. Please add $BIN_DIR to your PATH manually"
    fi
  fi
}

# Run the initialization
run_init() {
  if command -v "$BIN_DIR/echoy" >/dev/null 2>&1; then
    info "Start echoy..."
    "$BIN_DIR/echoy" help
  else
    warn "Could not run initialization. Please run 'echoy init' manually after restarting your terminal."
  fi
}

# Main installation process
main() {
  echo "============================================="
  echo "       Echoy Installation Script            "
  echo "============================================="

  detect_platform
  setup_directories
  download_binary
  create_symlink
  update_path

  success "Installation completed successfully!"
  success "Echoy has been installed to $INSTALL_DIR"

  if [ "$OS" = "windows" ]; then
    echo ""
    warn "Windows users: If you're not using Git Bash or WSL,"
    warn "please ensure $BIN_DIR is added to your system PATH."
  fi

  echo ""
  echo "To start using Echoy, you may need to:"
  echo "1. Restart your terminal or run 'source ~/.bashrc' (or equivalent)"
  echo "2. Run 'echoy init' to set up your configuration"
  echo "3. Start chatting with 'echoy chat'"
  echo ""

  run_init
}

# Execute the installation
main