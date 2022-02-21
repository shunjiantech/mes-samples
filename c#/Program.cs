// See https://aka.ms/new-console-template for more information

using System.Net.Http.Json;
using System.Security.Cryptography;
using System.Text;
using System.Web;

var baseUrl = "http://127.0.0.1:8080/api/open";
var appKey = "appKey";
var appSecret = "appSecret";

var sdk = new Sdk(baseUrl, appKey, appSecret);
sdk.GetDevices(67, "20210901103050484");
sdk.SaveTestData("20210901103050484", new Dictionary<string, object>[]
{
    new()
    {
        {"test_item_id", 10000},
        {
            "test_data", new Dictionary<string, object>[]
            {
                new()
                {
                    {"item1", 1},
                    {"item2", "2"},
                    {"item3", true},
                }
            }
        }
    }
});
sdk.UploadImage("../../../baidu.png");

class Sdk
{
    private readonly string _baseUrl;
    private readonly string _appKey;
    private readonly string _appSecret;

    private readonly HttpClient _client;

    public Sdk(string baseUrl, string appKey, string appSecret)
    {
        _baseUrl = baseUrl;
        _appKey = appKey;
        _appSecret = appSecret;
        _client = new HttpClient();
    }

    // 获取被试品信息
    public void GetDevices(long testStationId, string qrcode)
    {
        var request = new HttpRequestMessage(HttpMethod.Get,
            this._baseUrl + $"/devices?test_station_id={testStationId}&qrcode={qrcode}");

        SignRequest(request);

        var response = _client.SendAsync(request).Result;
        Console.WriteLine(response.Content.ReadAsStringAsync().Result);
    }

    // 保存试验数据
    public void SaveTestData(string qrcode, Dictionary<string, object>[] testData)
    {
        var content = JsonContent.Create(testData);

        var request = new HttpRequestMessage(HttpMethod.Post,
            this._baseUrl + $"/devices/{qrcode}/test_data");
        request.Content = content;

        SignRequest(request);

        var response = _client.SendAsync(request).Result;
        Console.WriteLine(response.Content.ReadAsStringAsync().Result);
    }

    // 上传试验图片
    public void UploadImage(string imagePath)
    {
        var stream = File.OpenRead(imagePath);
        var content = new MultipartFormDataContent();
        content.Add(new StreamContent(stream), "file", Path.GetFileName(imagePath));

        var request = new HttpRequestMessage(HttpMethod.Post, this._baseUrl + "/uploads");
        request.Content = content;

        SignRequest(request);

        var response = _client.SendAsync(request).Result;
        Console.WriteLine(response.Content.ReadAsStringAsync().Result);
    }

    // 签名
    private void SignRequest(HttpRequestMessage request)
    {
        var query = "";
        if (request.RequestUri?.Query != null)
        {
            var queryDict = HttpUtility.ParseQueryString(request.RequestUri.Query);
            var sortedDict =
                new SortedDictionary<string, string>(queryDict.AllKeys.ToDictionary(k => k, k => queryDict[k]));
            query = string.Join("&", sortedDict.Select(x => x.Key + "=" + x.Value).ToArray());
        }

        var contentMd5 = "";
        if (request.Content != null)
        {
            var body = request.Content.ReadAsByteArrayAsync().Result;
            contentMd5 = Md5(body);
        }

        var timestamp = new DateTimeOffset(DateTime.UtcNow).ToUnixTimeMilliseconds().ToString();
        var stringToSign = request.Method + "\n" +
                           request.RequestUri?.AbsolutePath + "\n" +
                           query + "\n" +
                           request.Content?.Headers.ContentType + "\n" +
                           contentMd5 + "\n" +
                           timestamp + "\n" +
                           _appSecret;
        var signature = HmacSha1(Encoding.UTF8.GetBytes(_appKey), Encoding.UTF8.GetBytes(stringToSign));

        request.Headers.Add("X-AppKey", _appKey);
        request.Headers.Add("X-TimeStamp", timestamp);
        request.Headers.Add("X-Signature", signature);
    }

    private string Md5(byte[] data)
    {
        var md5 = MD5.Create();
        var bytes = md5.ComputeHash(data);

        return Convert.ToBase64String(bytes);
    }

    private string HmacSha1(byte[] key, byte[] data)
    {
        var hmacSha1 = new HMACSHA1(key);
        var bytes = hmacSha1.ComputeHash(data);

        return Convert.ToBase64String(bytes);
    }
}