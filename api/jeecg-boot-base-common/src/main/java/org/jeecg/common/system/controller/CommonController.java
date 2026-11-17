package org.jeecg.common.system.controller;

import com.alibaba.fastjson.JSON;
import com.alibaba.fastjson.JSONObject;
import lombok.extern.slf4j.Slf4j;
import org.jeecg.common.api.vo.Result;
import org.jeecg.common.constant.CommonConstant;
import org.jeecg.common.system.api.ISysBaseAPI;
import org.jeecg.common.system.util.JwtUtil;
import org.jeecg.common.util.CommonUtils;
import org.jeecg.common.util.RestUtil;
import org.jeecg.common.util.TokenUtils;
import org.jeecg.common.util.oConvertUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpMethod;
import org.springframework.http.ResponseEntity;
import org.springframework.http.server.ServletServerHttpRequest;
import org.springframework.util.AntPathMatcher;
import org.springframework.util.FileCopyUtils;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;
import org.springframework.web.multipart.MultipartHttpServletRequest;
import org.springframework.web.servlet.HandlerMapping;
import org.springframework.web.servlet.ModelAndView;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.*;
import java.net.URLDecoder;

/**
 * <p>
 * 用户表 前端控制器
 * </p>
 *
 * @Author scott
 * @since 2018-12-20
 */
@Slf4j
@RestController
@RequestMapping("/sys/common")
public class CommonController {

	@Autowired
	private ISysBaseAPI sysBaseAPI;

	@Value(value = "${jeecg.path.upload}")
	private String uploadpath;

	/**
	 * 本地：local minio：minio 阿里：alioss
	 */
	@Value(value="${jeecg.uploadType}")
	private String uploadType;

	/**
	 * @Author 政辉
	 * @return
	 */
	@GetMapping("/403")
	public Result<?> noauth()  {
		return Result.error("没有权限，请联系管理员授权");
	}

	/**
	 * 文件上传统一方法
	 * @param request
	 * @param response
	 * @return
	 */
	@PostMapping(value = "/upload")
	public Result<?> upload(HttpServletRequest request, HttpServletResponse response) {
		Result<?> result = new Result<>();
		String savePath = "";
		String bizPath = request.getParameter("biz");
		MultipartHttpServletRequest multipartRequest = (MultipartHttpServletRequest) request;
		MultipartFile file = multipartRequest.getFile("file");// 获取上传文件对象
		if(oConvertUtils.isEmpty(bizPath)){
			if(CommonConstant.UPLOAD_TYPE_OSS.equals(uploadType)){
				//未指定目录，则用阿里云默认目录 upload
				bizPath = "upload";
				//result.setMessage("使用阿里云文件上传时，必须添加目录！");
				//result.setSuccess(false);
				//return result;
			}else{
				bizPath = "";
			}
		}
		if(CommonConstant.UPLOAD_TYPE_LOCAL.equals(uploadType)){
			//针对jeditor编辑器如何使 lcaol模式，采用 base64格式存储
			String jeditor = request.getParameter("jeditor");
			if(oConvertUtils.isNotEmpty(jeditor)){
				result.setMessage(CommonConstant.UPLOAD_TYPE_LOCAL);
				result.setSuccess(true);
				return result;
			}else{
				savePath = this.uploadLocal(file,bizPath);
			}
		}else{
			savePath = sysBaseAPI.upload(file,bizPath,uploadType);
		}
		if(oConvertUtils.isNotEmpty(savePath)){
			result.setMessage(savePath);
			result.setSuccess(true);
		}else {
			result.setMessage("上传失败！");
			result.setSuccess(false);
		}
		return result;
	}

	/**
	 * 本地文件上传
	 * @param mf 文件
	 * @param bizPath  自定义路径
	 * @return
	 */
	private String uploadLocal(MultipartFile mf,String bizPath){
		try {
			String ctxPath = uploadpath;
			String fileName = null;
			File file = new File(ctxPath + File.separator + bizPath + File.separator );
			if (!file.exists()) {
				file.mkdirs();// 创建文件根目录
			}
			String orgName = mf.getOriginalFilename();// 获取文件名
			orgName = CommonUtils.getFileName(orgName);
			if(orgName.indexOf(".")!=-1){
				fileName = orgName.substring(0, orgName.lastIndexOf(".")) + "_" + System.currentTimeMillis() + orgName.substring(orgName.indexOf("."));
			}else{
				fileName = orgName+ "_" + System.currentTimeMillis();
			}
			String savePath = file.getPath() + File.separator + fileName;
			File savefile = new File(savePath);
			FileCopyUtils.copy(mf.getBytes(), savefile);
			String dbpath = null;
			if(oConvertUtils.isNotEmpty(bizPath)){
				dbpath = bizPath + File.separator + fileName;
			}else{
				dbpath = fileName;
			}
			if (dbpath.contains("\\")) {
				dbpath = dbpath.replace("\\", "/");
			}
			return dbpath;
		} catch (IOException e) {
			log.error(e.getMessage(), e);
		}
		return "";
	}

//	@PostMapping(value = "/upload2")
//	public Result<?> upload2(HttpServletRequest request, HttpServletResponse response) {
//		Result<?> result = new Result<>();
//		try {
//			String ctxPath = uploadpath;
//			String fileName = null;
//			String bizPath = "files";
//			String tempBizPath = request.getParameter("biz");
//			if(oConvertUtils.isNotEmpty(tempBizPath)){
//				bizPath = tempBizPath;
//			}
//			String nowday = new SimpleDateFormat("yyyyMMdd").format(new Date());
//			File file = new File(ctxPath + File.separator + bizPath + File.separator + nowday);
//			if (!file.exists()) {
//				file.mkdirs();// 创建文件根目录
//			}
//			MultipartHttpServletRequest multipartRequest = (MultipartHttpServletRequest) request;
//			MultipartFile mf = multipartRequest.getFile("file");// 获取上传文件对象
//			String orgName = mf.getOriginalFilename();// 获取文件名
//			fileName = orgName.substring(0, orgName.lastIndexOf(".")) + "_" + System.currentTimeMillis() + orgName.substring(orgName.indexOf("."));
//			String savePath = file.getPath() + File.separator + fileName;
//			File savefile = new File(savePath);
//			FileCopyUtils.copy(mf.getBytes(), savefile);
//			String dbpath = bizPath + File.separator + nowday + File.separator + fileName;
//			if (dbpath.contains("\\")) {
//				dbpath = dbpath.replace("\\", "/");
//			}
//			result.setMessage(dbpath);
//			result.setSuccess(true);
//		} catch (IOException e) {
//			result.setSuccess(false);
//			result.setMessage(e.getMessage());
//			log.error(e.getMessage(), e);
//		}
//		return result;
//	}

	/**
	 * 预览图片&下载文件
	 * 请求地址：http://localhost:8080/common/static/{user/20190119/e1fe9925bc315c60addea1b98eb1cb1349547719_1547866868179.jpg}
	 *
	 * @param request
	 * @param response
	 */
	@GetMapping(value = "/static/**")
	public void view(HttpServletRequest request, HttpServletResponse response) {
		// ISO-8859-1 ==> UTF-8 进行编码转换
		String imgPath = extractPathFromPattern(request);
		if(oConvertUtils.isEmpty(imgPath) || imgPath=="null"){
			return;
		}
		// 其余处理略
		InputStream inputStream = null;
		OutputStream outputStream = null;
		try {
			imgPath = imgPath.replace("..", "");
			if (imgPath.endsWith(",")) {
				imgPath = imgPath.substring(0, imgPath.length() - 1);
			}
			String filePath = uploadpath + File.separator + imgPath;
			File file = new File(filePath);
			if(!file.exists()){
				response.setStatus(404);
				throw new RuntimeException("文件不存在..");
			}
			response.setContentType("application/force-download");// 设置强制下载不打开
			response.addHeader("Content-Disposition", "attachment;fileName=" + new String(file.getName().getBytes("UTF-8"),"iso-8859-1"));
			inputStream = new BufferedInputStream(new FileInputStream(filePath));
			outputStream = response.getOutputStream();
			byte[] buf = new byte[1024];
			int len;
			while ((len = inputStream.read(buf)) > 0) {
				outputStream.write(buf, 0, len);
			}
			response.flushBuffer();
		} catch (IOException e) {
			log.error("预览文件失败" + e.getMessage());
			response.setStatus(404);
			e.printStackTrace();
		} finally {
			if (inputStream != null) {
				try {
					inputStream.close();
				} catch (IOException e) {
					log.error(e.getMessage(), e);
				}
			}
			if (outputStream != null) {
				try {
					outputStream.close();
				} catch (IOException e) {
					log.error(e.getMessage(), e);
				}
			}
		}

	}

//	/**
//	 * 下载文件
//	 * 请求地址：http://localhost:8080/common/download/{user/20190119/e1fe9925bc315c60addea1b98eb1cb1349547719_1547866868179.jpg}
//	 *
//	 * @param request
//	 * @param response
//	 * @throws Exception
//	 */
//	@GetMapping(value = "/download/**")
//	public void download(HttpServletRequest request, HttpServletResponse response) throws Exception {
//		// ISO-8859-1 ==> UTF-8 进行编码转换
//		String filePath = extractPathFromPattern(request);
//		// 其余处理略
//		InputStream inputStream = null;
//		OutputStream outputStream = null;
//		try {
//			filePath = filePath.replace("..", "");
//			if (filePath.endsWith(",")) {
//				filePath = filePath.substring(0, filePath.length() - 1);
//			}
//			String localPath = uploadpath;
//			String downloadFilePath = localPath + File.separator + filePath;
//			File file = new File(downloadFilePath);
//	         if (file.exists()) {
//	         	response.setContentType("application/force-download");// 设置强制下载不打开            
//	 			response.addHeader("Content-Disposition", "attachment;fileName=" + new String(file.getName().getBytes("UTF-8"),"iso-8859-1"));
//	 			inputStream = new BufferedInputStream(new FileInputStream(file));
//	 			outputStream = response.getOutputStream();
//	 			byte[] buf = new byte[1024];
//	 			int len;
//	 			while ((len = inputStream.read(buf)) > 0) {
//	 				outputStream.write(buf, 0, len);
//	 			}
//	 			response.flushBuffer();
//	         }
//
//		} catch (Exception e) {
//			log.info("文件下载失败" + e.getMessage());
//			// e.printStackTrace();
//		} finally {
//			if (inputStream != null) {
//				try {
//					inputStream.close();
//				} catch (IOException e) {
//					e.printStackTrace();
//				}
//			}
//			if (outputStream != null) {
//				try {
//					outputStream.close();
//				} catch (IOException e) {
//					e.printStackTrace();
//				}
//			}
//		}
//
//	}

	/**
	 * @功能：pdf预览Iframe
	 * @param modelAndView
	 * @return
	 */
	@RequestMapping("/pdf/pdfPreviewIframe")
	public ModelAndView pdfPreviewIframe(ModelAndView modelAndView) {
		modelAndView.setViewName("pdfPreviewIframe");
		return modelAndView;
	}

	/**
	  *  把指定URL后的字符串全部截断当成参数
	  *  这么做是为了防止URL中包含中文或者特殊字符（/等）时，匹配不了的问题
	 * @param request
	 * @return
	 */
	private static String extractPathFromPattern(final HttpServletRequest request) {
		String path = (String) request.getAttribute(HandlerMapping.PATH_WITHIN_HANDLER_MAPPING_ATTRIBUTE);
		String bestMatchPattern = (String) request.getAttribute(HandlerMapping.BEST_MATCHING_PATTERN_ATTRIBUTE);
		return new AntPathMatcher().extractPathWithinPattern(bestMatchPattern, path);
	}

    /**
     * 中转HTTP请求，解决跨域问题
     *
     * @param url 必填：请求地址
     * @return
     */
    @RequestMapping("/transitRESTful")
    public Result transitRESTful(@RequestParam("url") String url, HttpServletRequest request) {
        // 安全加固(V2.6)：本接口会把请求转发到 url 指定的任意地址，
        // 属于典型的 SSRF（服务端请求伪造）入口，可被用来探测内网、读取云元数据
        // (169.254.169.254)、访问 Redis/MySQL 等仅内网可达的服务。
        // 前端代码中未检索到对该接口的调用，这里做严格出站校验后仍保留接口。
        String urlCheckError = validateOutboundUrl(url);
        if (urlCheckError != null) {
            log.warn("transitRESTful 出站地址被拒绝: {} ({})", url, urlCheckError);
            return Result.error("非法的中转地址: " + urlCheckError);
        }
        try {
            ServletServerHttpRequest httpRequest = new ServletServerHttpRequest(request);
            // 中转请求method、body
            HttpMethod method = httpRequest.getMethod();
            JSONObject params;
            try {
                params = JSON.parseObject(JSON.toJSONString(httpRequest.getBody()));
            } catch (Exception e) {
                params = new JSONObject();
            }
            // 中转请求问号参数
            JSONObject variables = JSON.parseObject(JSON.toJSONString(request.getParameterMap()));
            variables.remove("url");
            // 在 headers 里传递Token
            String token = TokenUtils.getTokenByRequest(request);
            HttpHeaders headers = new HttpHeaders();
            headers.set("X-Access-Token", token);
            // 发送请求
            String httpURL = URLDecoder.decode(url, "UTF-8");
            ResponseEntity<String> response = RestUtil.request(httpURL, method, headers , variables, params, String.class);
            // 封装返回结果
            Result<Object> result = new Result<>();
            int statusCode = response.getStatusCodeValue();
            result.setCode(statusCode);
            result.setSuccess(statusCode == 200);
            String responseBody = response.getBody();
            try {
                // 尝试将返回结果转为JSON
                Object json = JSON.parse(responseBody);
                result.setResult(json);
            } catch (Exception e) {
                // 转成JSON失败，直接返回原始数据
                result.setResult(responseBody);
            }
            return result;
        } catch (Exception e) {
            log.debug("中转HTTP请求失败", e);
            return Result.error(e.getMessage());
        }
    }

    /**
     * 校验中转目标地址是否允许出站，用于阻断 SSRF。
     *
     * <p>规则：
     * <ul>
     *   <li>只允许 http / https；</li>
     *   <li>拒绝回环地址、链路本地地址（含 169.254.169.254 云元数据）、
     *       RFC1918 私有网段、CGNAT、组播与保留地址；</li>
     *   <li>拒绝裸 IP 形式的内部主机名写法与 localhost 别名。</li>
     * </ul>
     * 注意：DNS 解析后仍可能发生 rebinding，生产环境建议再叠加出站防火墙白名单。
     *
     * @param rawUrl 待校验地址
     * @return null 表示通过；否则返回拒绝原因
     */
    private String validateOutboundUrl(String rawUrl) {
        if (oConvertUtils.isEmpty(rawUrl)) {
            return "地址为空";
        }
        java.net.URI uri;
        try {
            uri = new java.net.URI(rawUrl.trim());
        } catch (Exception e) {
            return "地址格式不合法";
        }
        String scheme = uri.getScheme();
        if (scheme == null || !("http".equalsIgnoreCase(scheme) || "https".equalsIgnoreCase(scheme))) {
            return "仅允许 http/https 协议";
        }
        String host = uri.getHost();
        if (oConvertUtils.isEmpty(host)) {
            return "缺少主机名";
        }
        String lowerHost = host.toLowerCase();
        if ("localhost".equals(lowerHost) || lowerHost.endsWith(".localhost") || lowerHost.endsWith(".internal")) {
            return "禁止访问本机或内部域名";
        }
        try {
            java.net.InetAddress[] addresses = java.net.InetAddress.getAllByName(host);
            for (java.net.InetAddress address : addresses) {
                if (address.isLoopbackAddress()
                        || address.isLinkLocalAddress()
                        || address.isSiteLocalAddress()
                        || address.isAnyLocalAddress()
                        || address.isMulticastAddress()) {
                    return "禁止访问内网/保留地址";
                }
                byte[] raw = address.getAddress();
                // 100.64.0.0/10 (CGNAT)
                if (raw.length == 4 && (raw[0] & 0xFF) == 100 && ((raw[1] & 0xFF) >= 64 && (raw[1] & 0xFF) <= 127)) {
                    return "禁止访问运营商级 NAT 地址";
                }
                // 0.0.0.0/8
                if (raw.length == 4 && (raw[0] & 0xFF) == 0) {
                    return "禁止访问保留地址";
                }
                // IPv6 唯一本地地址 fc00::/7
                if (raw.length == 16 && (raw[0] & 0xFE) == 0xFC) {
                    return "禁止访问 IPv6 唯一本地地址";
                }
            }
        } catch (Exception e) {
            return "主机名无法解析";
        }
        return null;
    }

}
