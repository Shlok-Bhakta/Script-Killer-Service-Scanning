from flask import Flask, request
import yaml
import requests

app = Flask(__name__)

@app.route('/config', methods=['POST'])
def load_config():
    config_data = request.data
    config = yaml.load(config_data)
    return config

@app.route('/fetch')
def fetch_data():
    url = request.args.get('url')
    response = requests.get(url, verify=False)
    return response.text

if __name__ == '__main__':
    app.run(debug=True)
