package mes_samples;

import com.fasterxml.jackson.databind.ObjectMapper;
import okhttp3.*;
import okio.Buffer;

import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import java.io.File;
import java.security.InvalidKeyException;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.util.*;

public class Demo {
    public static void main(String[] args) {
        String baseUrl = "http://127.0.0.1:8080/api/open";
        String appKey = "appKey";
        String appSecret = "appSecret";

        Sdk sdk = new Sdk(baseUrl, appKey, appSecret);
        try {
            sdk.GetDevices(67, "20210901103050484");
            sdk.SaveTestData("20210901103050484", new ArrayList<Map<String, Object>>() {
                {
                    add(new HashMap<String, Object>() {
                        {
                            put("test_item_id", 10000);
                            put("test_data", new ArrayList<Map<String, Object>>() {
                                {
                                    add(new HashMap<String, Object>() {
                                        {
                                            put("item1", 1);
                                            put("item2", "2");
                                            put("item3", true);
                                        }
                                    });
                                }
                            });
                        }
                    });
                }
            });
            sdk.UploadImage("baidu.png");
            sdk.PingInstrument(100);
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    static class Sdk {
        private final OkHttpClient client;
        private final String baseUrl;
        private final String appKey;
        private final String appSecret;

        public Sdk(String baseUrl, String appKey, String appSecret) {
            this.baseUrl = baseUrl;
            this.appKey = appKey;
            this.appSecret = appSecret;
            this.client = new OkHttpClient();
        }

        // 获取被试品信息
        public void GetDevices(long testStationId, String qrcode) throws Exception {
            Request request = new Request.Builder()
                    .url(String.format("%s/devices?test_station_id=%d&qrcode=%s", baseUrl, testStationId, qrcode))
                    .build();

            request = SignRequest(request);

            Response response = client.newCall(request).execute();

            System.out.println(response.body().string());
        }

        // 保存试验数据
        public void SaveTestData(String qrcode, List<Map<String, Object>> data) throws Exception {
            String json = new ObjectMapper().writeValueAsString(data);
            RequestBody body = RequestBody.create(json, MediaType.parse("application/json; charset=utf-8"));

            Request request = new Request.Builder()
                    .url(String.format("%s/devices/%s/test_data", baseUrl, qrcode))
                    .post(body)
                    .build();

            request = SignRequest(request);

            Response response = client.newCall(request).execute();

            System.out.println(response.body().string());
        }

        // 上传试验图片
        public void UploadImage(String imagePath) throws Exception {
            File file = new File(imagePath);
            RequestBody fileBody = RequestBody.create(file, MediaType.parse("image/png"));
            RequestBody body = new MultipartBody.Builder()
                    .setType(MultipartBody.FORM)
                    .addFormDataPart("file", file.getName(), fileBody)
                    .build();

            Request request = new Request.Builder()
                    .url(String.format("%s/uploads", baseUrl))
                    .post(body)
                    .build();

            request = SignRequest(request);

            Response response = client.newCall(request).execute();

            System.out.println(response.body().string());
        }

        // 心跳
        public void PingInstrument(long instrumentId) throws Exception {
            RequestBody body = RequestBody.create(null, new byte[0]);

            Request request = new Request.Builder()
                    .url(String.format("%s/instruments/%d/ping", baseUrl, instrumentId))
                    .post(body)
                    .build();

            request = SignRequest(request);

            Response response = client.newCall(request).execute();

            System.out.println(response.body().string());
        }

        private Request SignRequest(Request request) throws Exception {
            List<String> queryList = new ArrayList<>();
            Set<String> querySet = request.url().queryParameterNames();
            if (querySet.size() > 0) {
                Set<String> sortedSet = new TreeSet<>(querySet);
                for (String key : sortedSet) {
                    queryList.add(String.format("%s=%s", key, request.url().queryParameter(key)));
                }
            }
            String query = String.join("&", queryList);

            String contentType = "";
            String contentMd5 = "";
            if (request.body() != null && request.body().contentLength() > 0) {
                contentType = request.body().contentType().toString();

                Buffer buffer = new Buffer();
                request.body().writeTo(buffer);

                contentMd5 = md5(buffer.readByteArray());
            }

            long timestamp = System.currentTimeMillis();
            String stringToSign = request.method() + "\n" +
                    request.url().url().getPath() + "\n" +
                    query + "\n" +
                    contentType + "\n" +
                    contentMd5 + "\n" +
                    timestamp + "\n" +
                    appSecret;
            String signature = hmacSha1(appKey.getBytes(), stringToSign.getBytes());

            return request.newBuilder()
                    .header("X-AppKey", appKey)
                    .header("X-Timestamp", String.valueOf(timestamp))
                    .header("X-Signature", signature)
                    .build();
        }

        private String md5(byte[] data) throws NoSuchAlgorithmException {
            MessageDigest md5 = MessageDigest.getInstance("MD5");
            md5.update(data);

            byte[] digest = md5.digest();

            return Base64.getEncoder().encodeToString(digest);
        }

        private String hmacSha1(byte[] key, byte[] data) throws NoSuchAlgorithmException, InvalidKeyException {
            SecretKeySpec signKey = new SecretKeySpec(key, "HmacSHA1");

            Mac mac = Mac.getInstance("HmacSHA1");
            mac.init(signKey);
            mac.update(data);

            byte[] digest = mac.doFinal();

            return Base64.getEncoder().encodeToString(digest);
        }
    }
}
