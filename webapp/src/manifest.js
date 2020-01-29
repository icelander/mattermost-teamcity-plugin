// This file is automatically generated. Do not modify it manually.

const manifest = JSON.parse(`
{
    "id": "mattermost-teamcity-plugin",
    "name": "Mattermost TeamCity Plugin",
    "description": "This plugin integrates Mattermost with TeamCity",
    "version": "1.0.1",
    "min_server_version": "5.18.0",
    "server": {
        "executables": {
            "linux-amd64": "server/dist/plugin-linux-amd64",
            "darwin-amd64": "server/dist/plugin-darwin-amd64",
            "windows-amd64": "server/dist/plugin-windows-amd64.exe"
        },
        "executable": ""
    },
    "webapp": {
        "bundle_path": "webapp/dist/main.js"
    },
    "settings_schema": {
        "header": "**Note:** Because it uses token-based authentication this plugin supports TeamCity 2019.1 and higher",
        "footer": "",
        "settings": [
            {
                "key": "TeamCityURL",
                "display_name": "TeamCity URL",
                "type": "text",
                "help_text": "The URL of your TeamCity Server",
                "placeholder": "http://teamcity:8111/",
                "default": ""
            },
            {
                "key": "TeamCityToken",
                "display_name": "TeamCity Access Token",
                "type": "text",
                "help_text": "The access token to use. [See documentation for more information](https://www.jetbrains.com/help/teamcity/managing-your-user-account.html#ManagingyourUserAccount-ManagingAccessTokens)",
                "placeholder": "eyJ0eXAiOiAiVENWMiJ9.d21QeUw2akYwclFBQTVtUGlxY2xOWWV4TVNz.MDViNmM0Y2EtNzc5YS00MDU5LWE0NTgtYmVmNzg4YzhjMGVl",
                "default": ""
            },
            {
                "key": "TeamCityMaxBuilds",
                "display_name": "Builds Returned",
                "type": "text",
                "help_text": "Number of builds returned when listing builds",
                "placeholder": "5",
                "default": "5"
            }
        ]
    }
}
`);

export default manifest;
export const id = manifest.id;
export const version = manifest.version;
