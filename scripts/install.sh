if [ "$EUID" -ne 0 ]; then
    echo "Please run as root or sudoers."
    exit 1
fi

echo "Do you want to download RustnessCloud? (y/n)"
read wantInstall

if [ "$wantInstall" != "y" ]; then
    echo "Aborting..."
    exit 1
fi

# Function to install package if missing
install_package_if_missing() {
    local package="$1"
    hash "$package" 2>/dev/null || {
        echo >&2 "You need to install $package. Do you want to install it? (y/n)";
        read wantInstall
        if [ "$wantInstall" = "y" ]; then
            sudo apt install "$package" -y
        else
            echo "Aborting!"
            exit 1
        fi
    }
}

# Check and install required packages
install_package_if_missing "wget"
install_package_if_missing "unzip"
install_package_if_missing "screen"

# Check and install Java
hash java 2>/dev/null || {
    echo "Do you want to install java? (y/n)"

    read wantInstall
    if [ "$wantInstall" = "y" ]; then
        echo "Select a java version(Minecraft Version 1.7.10 - 1.16.5: Java 8 | 1.9 - 1.16.5: Java 11 | 1.17 - 1.17.1: Java 16 | 1.17 - 1.20.4: Java 17 | 1.20.5 - latest: Java 21)"
        read javaVersion
        sudo apt install -y software-properties-common
        sudo apt install -y wget apt-transport-https gpg
        sudo wget -qO - https://packages.adoptium.net/artifactory/api/gpg/key/public | sudo gpg --dearmor | sudo tee /etc/apt/trusted.gpg.d/adoptium.gpg > /dev/null
        if [ -f "/etc/os-release" ]; then
            sudo echo "deb https://packages.adoptium.net/artifactory/deb $(awk -F= '/^VERSION_CODENAME/{print$2}' /etc/os-release) main" | sudo tee /etc/apt/sources.list.d/adoptium.list
        else
          sudo echo "deb https://packages.adoptium.net/artifactory/deb $(awk -F= '/^VERSION_CODENAME/{print$2}' /etc/os-release) main" | sudo tee /etc/apt/sources.list.d/adoptium.list
        fi
          sudo apt install -y temurin-"$javaVersion"-jdk
    else
        echo "Aborting..."
        exit 1
    fi
}
hash java 2>/dev/null || {
    echo >&2 "You need to install Java. Do you want to install it? (y/n)";
    read wantInstall
    if [ "$wantInstall" = "y" ]; then
        curl https://sh.rustup.rs -sSf | sh
    else
        echo "Aborting..."
        exit 1
    fi
}

# Download and install RustnessCloud
echo "Downloading RustnessCloud..."
sleep 0.5

wget https://github.com/RustServices/RustnessCloud/releases/latest/download/rustnesscloud.zip

echo "Unzipping RustnessCloud..."
sleep 0.5

unzip rustnesscloud.zip
chmod a+x start.sh

echo "RustnessCloud installation completed successfully."

# Prompt to start RustnessCloud
echo "Do you want to start RustnessCloud? (y/n)"
read wantStart

if [ "$wantStart" = "y" ]; then
    bash start.sh
else
    echo "Exiting."
fi