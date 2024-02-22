#include <iostream>
#include <filesystem>
#include <string>
#include <fstream>

namespace fs = std::filesystem;

int iclangNum = 0;
long long frontTime = 0;
long long totalTime = 0;

void traverseDirectory(const fs::path& directoryPath) {
    for (const auto& entry : fs::directory_iterator(directoryPath)) {
        if (fs::is_directory(entry.path())) {
            if (entry.path().extension() == ".iclang") {
		std::filesystem::path compileTxtPath = entry.path() / "compile.txt";
		if (std::filesystem::exists(compileTxtPath)) {
		    iclangNum += 1;
		    std::ifstream inputFile(compileTxtPath);

		    if (!inputFile.is_open()) {
		        std::cerr << "cannot open file " << compileTxtPath << std::endl;
		    }

		    std::string temp;
		    std::getline(inputFile, temp);
		    std::getline(inputFile, temp);
		    std::getline(inputFile, temp);
		    long long ft, tt;
		    inputFile >> ft >> tt;
		    frontTime += ft;
		    totalTime += tt;

		    inputFile.close();
		}
                // std::cout << "Found directory with .iclang extension: " << entry.path() << std::endl;
            } else {
                traverseDirectory(entry.path());
            }
        }
    }
}

int main(int argc, char **argv) {
    if (argc != 2) {
	std::cerr << "Please provide a directory" << std::endl;
	return 1;
    }

    std::string rootPath = argv[1];
    fs::path rootDir(rootPath);

    if (!fs::exists(rootDir) || !fs::is_directory(rootDir)) {
        std::cerr << "Invalid directory: " << rootPath << std::endl;
        return 1;
    }

    traverseDirectory(rootDir);

    long long backTime = totalTime - frontTime;

    std::cout << "[iclangNum] " << iclangNum << std::endl;
    std::cout << "[totalTime] " << totalTime << std::endl;
    std::cout << "[frontTime] " << frontTime << " " << 1.0*frontTime/totalTime << std::endl;
    std::cout << "[backTime] " << backTime << " " << 1.0*backTime/totalTime << std::endl;

    return 0;
}
