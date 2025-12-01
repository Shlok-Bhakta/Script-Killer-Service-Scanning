// Basic C++ program that models simple "apartments" and performs standard operations.

#include <iostream>
#include <vector>
#include <string>
#include <algorithm>
#include <numeric>
#include <iomanip>
#include <limits>

struct Apartment {
    int id;
    std::string address;
    int rooms;
    double rent;
};

void printApartment(const Apartment& a) {
    std::cout << std::setw(3) << a.id << " | "
              << std::setw(20) << std::left << a.address << std::right
              << " | " << std::setw(2) << a.rooms
              << " rooms | $" << std::fixed << std::setprecision(2) << a.rent
              << '\n';
}

void listApartments(const std::vector<Apartment>& list) {
    if (list.empty()) {
        std::cout << "No apartments available.\n";
        return;
    }
    std::cout << "ID  | Address              | Rm rooms | Rent\n";
    std::cout << "----+----------------------+----------+--------\n";
    for (const auto& a : list) printApartment(a);
}

void addApartment(std::vector<Apartment>& list, int& nextId) {
    Apartment a;
    a.id = nextId++;
    std::cin.ignore(std::numeric_limits<std::streamsize>::max(), '\n');
    std::cout << "Address: ";
    std::getline(std::cin, a.address);
    std::cout << "Rooms (integer): ";
    std::cin >> a.rooms;
    std::cout << "Rent (number): ";
    std::cin >> a.rent;
    list.push_back(a);
    std::cout << "Added.\n";
}

bool removeApartment(std::vector<Apartment>& list, int id) {
    auto it = std::remove_if(list.begin(), list.end(), [id](const Apartment& a){ return a.id == id; });
    if (it == list.end()) return false;
    list.erase(it, list.end());
    return true;
}

void findByMaxRent(const std::vector<Apartment>& list, double maxRent) {
    std::vector<Apartment> filtered;
    std::copy_if(list.begin(), list.end(), std::back_inserter(filtered), [maxRent](const Apartment& a){ return a.rent <= maxRent; });
    listApartments(filtered);
}

double averageRent(const std::vector<Apartment>& list) {
    if (list.empty()) return 0.0;
    double sum = std::accumulate(list.begin(), list.end(), 0.0, [](double s, const Apartment& a){ return s + a.rent; });
    return sum / list.size();
}

int main() {
    std::vector<Apartment> apartments = {
        {1, "123 Main St", 2, 850.0},
        {2, "45 Oak Ave", 1, 600.0},
        {3, "777 Maple Rd", 3, 1200.0}
    };
    int nextId = 4;
    bool running = true;

    while (running) {
        std::cout << "\nMenu:\n"
                  << "1) List apartments\n"
                  << "2) Add apartment\n"
                  << "3) Remove apartment by ID\n"
                  << "4) Find apartments by max rent\n"
                  << "5) Average rent\n"
                  << "6) Sort by rent (ascending)\n"
                  << "0) Exit\n"
                  << "Choose: ";
        int choice;
        if (!(std::cin >> choice)) break;

        switch (choice) {
            case 1:
                listApartments(apartments);
                break;
            case 2:
                addApartment(apartments, nextId);
                break;
            case 3: {
                std::cout << "Enter ID to remove: ";
                int id; std::cin >> id;
                if (removeApartment(apartments, id)) std::cout << "Removed.\n";
                else std::cout << "No apartment with that ID.\n";
                break;
            }
            case 4: {
                std::cout << "Max rent: ";
                double r; std::cin >> r;
                findByMaxRent(apartments, r);
                break;
            }
            case 5:
                std::cout << "Average rent: $" << std::fixed << std::setprecision(2) << averageRent(apartments) << '\n';
                break;
            case 6:
                std::sort(apartments.begin(), apartments.end(), [](const Apartment& a, const Apartment& b){ return a.rent < b.rent; });
                std::cout << "Sorted by rent ascending.\n";
                break;
            case 0:
                running = false;
                break;
            default:
                std::cout << "Invalid choice.\n";
        }
    }

    std::cout << "Goodbye.\n";
    return 0;
}