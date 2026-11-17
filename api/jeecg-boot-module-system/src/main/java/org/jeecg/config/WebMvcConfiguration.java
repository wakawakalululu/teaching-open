package org.jeecg.config;

import org.jeecg.modules.shiro.authc.interceptor.OnlineInterceptor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.cors.CorsConfiguration;
import org.springframework.web.cors.UrlBasedCorsConfigurationSource;
import org.springframework.web.filter.CorsFilter;
import org.springframework.web.servlet.config.annotation.InterceptorRegistry;
import org.springframework.web.servlet.config.annotation.ResourceHandlerRegistry;
import org.springframework.web.servlet.config.annotation.ViewControllerRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

/**
 * Spring Boot 2.0 解决跨域问题
 *
 * @Author qinfeng
 *
 */
@Configuration
public class WebMvcConfiguration implements WebMvcConfigurer {

	@Value("${jeecg.path.upload}")
	private String upLoadPath;
	@Value("${jeecg.path.webapp}")
	private String webAppPath;
	@Value("${spring.resource.static-locations}")
	private String staticLocations;

	@Bean
	public OnlineInterceptor onlineInterceptor(){
		return new OnlineInterceptor();
	}

	/**
	 * 允许的跨域来源白名单，逗号分隔，由 jeecg.cors.allowedOrigins 配置。
	 *
	 * <p>安全加固(V2.6)：原实现为 setAllowCredentials(true) + addAllowedOrigin("*") 组合，
	 * 该组合会让任意站点携带用户 Cookie/凭据发起跨域请求并读取响应，
	 * 属于高风险配置（浏览器虽会拒绝 "*" 与 credentials 共存，但部分网关/旧客户端
	 * 仍可能出现绕过或降级行为）。现改为显式白名单，默认仅放行本站域名。
	 */
	@Value("${jeecg.cors.allowedOrigins:}")
	private String allowedOrigins;

	@Bean
	public CorsFilter corsFilter() {
		final UrlBasedCorsConfigurationSource urlBasedCorsConfigurationSource = new UrlBasedCorsConfigurationSource();
		final CorsConfiguration corsConfiguration = new CorsConfiguration();
		/* 是否允许请求带有验证信息 */
		corsConfiguration.setAllowCredentials(true);
		/* 允许访问的客户端域名：仅白名单，禁止使用 "*" */
		if (allowedOrigins == null || allowedOrigins.trim().isEmpty()) {
			// 未配置时退化为仅允许同源（不添加任何 Origin，等价于不放行跨域）
			corsConfiguration.setAllowedOrigins(new java.util.ArrayList<String>());
		} else {
			for (String origin : allowedOrigins.split(",")) {
				String item = origin.trim();
				if (!item.isEmpty() && !"*".equals(item)) {
					corsConfiguration.addAllowedOrigin(item);
				}
			}
		}
		/* 允许服务端访问的客户端请求头 */
		corsConfiguration.addAllowedHeader("*");
		/* 允许访问的方法名,GET POST等（不放开 TRACE/TRACK） */
		corsConfiguration.addAllowedMethod("GET");
		corsConfiguration.addAllowedMethod("POST");
		corsConfiguration.addAllowedMethod("PUT");
		corsConfiguration.addAllowedMethod("DELETE");
		corsConfiguration.addAllowedMethod("OPTIONS");
		corsConfiguration.setMaxAge(3600L);
		urlBasedCorsConfigurationSource.registerCorsConfiguration("/**", corsConfiguration);
		return new CorsFilter(urlBasedCorsConfigurationSource);
	}

	/**
	 * 静态资源的配置 - 使得可以从磁盘中读取 Html、图片、视频、音频等
	 */
	@Override
	public void addResourceHandlers(ResourceHandlerRegistry registry) {
		registry.addResourceHandler("/**")
		.addResourceLocations("file:" + upLoadPath + "//", "file:" + webAppPath + "//")
		.addResourceLocations(staticLocations.split(","));
	}

	/**
	 * 方案一： 默认访问根路径跳转 doc.html页面 （swagger文档页面）
	 * 方案二： 访问根路径改成跳转 index.html页面 （简化部署方案： 可以把前端打包直接放到项目的 webapp，上面的配置）
	 */
	@Override
	public void addViewControllers(ViewControllerRegistry registry) {
		registry.addViewController("/").setViewName("doc.html");
	}

	@Override
	public void addInterceptors(InterceptorRegistry registry) {
		String [] exculudes = new String[]{"/*.html","/html/**","/js/**","/css/**","/images/**"};
		registry.addInterceptor(onlineInterceptor()).excludePathPatterns(exculudes).addPathPatterns("/online/cgform/api/**");
	}
}
