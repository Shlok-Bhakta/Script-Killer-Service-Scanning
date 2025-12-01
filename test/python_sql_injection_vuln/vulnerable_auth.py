import sqlite3

def get_connection():
    return sqlite3.connect('test.db')

def login(username, password):
    conn = get_connection()
    cursor = conn.cursor()
    
    query = "SELECT * FROM users WHERE username='" + username + "' AND password='" + password + "'"
    print(f"Executing query: {query}")
    
    cursor.execute(query)
    user = cursor.fetchone()
    conn.close()
    
    if user:
        return {"id": user[0], "username": user[1], "email": user[3], "is_admin": user[4]}
    return None

def get_user_by_username(username):
    conn = get_connection()
    cursor = conn.cursor()
    
    query = f"SELECT * FROM users WHERE username='{username}'"
    print(f"Executing query: {query}")
    
    cursor.execute(query)
    user = cursor.fetchone()
    conn.close()
    
    return user

def search_users(search_term):
    conn = get_connection()
    cursor = conn.cursor()
    
    query = "SELECT username, email FROM users WHERE username LIKE '%" + search_term + "%'"
    print(f"Executing query: {query}")
    
    cursor.execute(query)
    users = cursor.fetchall()
    conn.close()
    
    return users

if __name__ == '__main__':
    from setup_db import setup_database
    setup_database()
    
    print("\n--- Normal login ---")
    result = login("admin", "supersecret123")
    print(f"Result: {result}")
    
    print("\n--- SQL Injection: bypass auth ---")
    result = login("admin", "' OR '1'='1")
    print(f"Result: {result}")
    
    print("\n--- SQL Injection: login as admin without password ---")
    result = login("admin'--", "anything")
    print(f"Result: {result}")
    
    print("\n--- SQL Injection: dump all users ---")
    result = search_users("' OR '1'='1")
    print(f"Result: {result}")
