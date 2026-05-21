import subprocess
import time
import urllib.request
import urllib.parse
import http.cookiejar
import sys
import os

PORT = "30012"
BASE_URL = f"http://localhost:{PORT}"

def run_server():
    print("Starting Go server...")
    env = os.environ.copy()
    env["PORT"] = PORT
    env["DATABASE_URL"] = "test_smoke_database.db"
    # Remove old test DB if exists
    if os.path.exists("test_smoke_database.db"):
        os.remove("test_smoke_database.db")
        
    process = subprocess.Popen(
        ["go", "run", "cmd/server/main.go"],
        cwd="go-backend",
        env=env
    )
    return process

def wait_for_server():
    retries = 10
    while retries > 0:
        try:
            with urllib.request.urlopen(BASE_URL) as response:
                if response.status == 200:
                    print("Server is up and running!")
                    return True
        except Exception:
            pass
        time.sleep(1)
        retries -= 1
    return False

def test_smoke():
    # Setup Cookie Jar to handle session cookies
    cj = http.cookiejar.CookieJar()
    opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(cj))
    urllib.request.install_opener(opener)

    print("\n--- Running Smoke Tests ---\n")

    # 1. Main Page
    print("1. Testing Main page...")
    with urllib.request.urlopen(BASE_URL) as response:
        html = response.read().decode('utf-8')
        assert "Dragon of Icespire Peak" in html, "Seed game not found on homepage"
        print("   -> Success: Seed game list display verified.")

    # 2. Login as GM
    print("2. Testing GM Login...")
    login_data = urllib.parse.urlencode({
        'username': 'gm',
        'password': 'password'
    }).encode('utf-8')
    
    req = urllib.request.Request(f"{BASE_URL}/login", data=login_data, method='POST')
    with urllib.request.urlopen(req) as response:
        html = response.read().decode('utf-8')
        assert "GM Dashboard" in html, "GM Dashboard not reached after login"
        print("   -> Success: GM logged in and dashboard reached.")

    # 3. Create a Game as GM
    print("3. Testing Game Creation...")
    create_data = urllib.parse.urlencode({
        'title': 'Smoke Test Campaign',
        'system': 'Pathfinder 2e',
        'start_time': '2026-06-01T18:00',
        'price': '20.00',
        'max_players': '6',
        'description': 'A high-stakes pathfinder experience.'
    }).encode('utf-8')
    req = urllib.request.Request(f"{BASE_URL}/gm_dashboard/create_game", data=create_data, method='POST')
    with urllib.request.urlopen(req) as response:
        html = response.read().decode('utf-8')
        assert "Smoke Test Campaign" in html, "Created game title not found on GM Dashboard"
        print("   -> Success: Game successfully created and listed.")

    # 4. Logout GM
    print("4. Testing GM Logout...")
    req = urllib.request.Request(f"{BASE_URL}/logout", data=b'', method='POST')
    with urllib.request.urlopen(req) as response:
        html = response.read().decode('utf-8')
        assert "Dragon of Icespire Peak" in html, "Logout did not redirect back to front page"
        print("   -> Success: GM logged out.")

    # 5. Register & Login a new Player
    print("5. Registering new player 'smoke_player'...")
    reg_data = urllib.parse.urlencode({
        'username': 'smoke_player',
        'password': 'password',
        'role': 'Player'
    }).encode('utf-8')
    req = urllib.request.Request(f"{BASE_URL}/register", data=reg_data, method='POST')
    with urllib.request.urlopen(req) as response:
        assert response.status == 200, "Registration did not return success"
        print("   -> Success: Player registered.")

    print("6. Logging in player 'smoke_player'...")
    player_login = urllib.parse.urlencode({
        'username': 'smoke_player',
        'password': 'password'
    }).encode('utf-8')
    req = urllib.request.Request(f"{BASE_URL}/login", data=player_login, method='POST')
    with urllib.request.urlopen(req) as response:
        html = response.read().decode('utf-8')
        assert "Player Dashboard" in html, "Player Dashboard not reached after login"
        print("   -> Success: Player logged in and dashboard reached.")

    # 6. Join the seeded game (ID 1)
    print("7. Player joining seeded game...")
    req = urllib.request.Request(f"{BASE_URL}/game/1/join", data=b'', method='POST')
    with urllib.request.urlopen(req) as response:
        html = response.read().decode('utf-8')
        assert "Player Dashboard" in html, "Did not return to Player Dashboard after joining"
        print("   -> Success: Join request submitted successfully.")

    print("\nAll Smoke Tests Passed Successfully!")

if __name__ == "__main__":
    process = run_server()
    try:
        if wait_for_server():
            test_smoke()
        else:
            print("Server failed to respond in time.")
            sys.exit(1)
    except Exception as e:
        print(f"Error during smoke test execution: {e}")
        # Gather logs/output from process if possible, but keep it minimal
        sys.exit(1)
    finally:
        print("Shutting down the server...")
        process.terminate()
        process.wait()
        # Clean up test DB
        if os.path.exists("go-backend/test_smoke_database.db"):
            os.remove("go-backend/test_smoke_database.db")
        print("Cleaned up.")
