package com.training.server.controller;

import com.training.common.dto.request.CreateCourseRequest;
import com.training.common.dto.request.UpdateCourseRequest;
import com.training.common.response.ApiResponse;
import com.training.common.dto.response.CourseDTO;
import com.training.common.enums.ErrorCode;
import com.training.server.service.CourseService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping("/api/courses")
public class CourseController {

    @Autowired
    private CourseService courseService;

    @PostMapping
    public ApiResponse<CourseDTO> createCourse(@RequestBody CreateCourseRequest request) {
        try {
            CourseDTO course = courseService.createCourse(request);
            return ApiResponse.success(course);
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.INVALID_PARAM, e.getMessage());
        }
    }

    @GetMapping("/{id}")
    public ApiResponse<CourseDTO> getCourseById(@PathVariable String id) {
        try {
            CourseDTO course = courseService.getCourseById(id);
            return ApiResponse.success(course);
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.COURSE_NOT_FOUND);
        }
    }

    @GetMapping
    public ApiResponse<List<CourseDTO>> getAllCourses(@RequestParam(required = false) Boolean publishedOnly) {
        List<CourseDTO> courses;
        if (Boolean.TRUE.equals(publishedOnly)) {
            courses = courseService.getPublishedCourses();
        } else {
            courses = courseService.getAllCourses();
        }
        return ApiResponse.success(courses);
    }

    @GetMapping(params = "instructorId")
    public ApiResponse<List<CourseDTO>> getCoursesByInstructor(@RequestParam String instructorId) {
        try {
            List<CourseDTO> courses = courseService.getCoursesByInstructor(instructorId);
            return ApiResponse.success(courses);
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.INSTRUCTOR_NOT_FOUND);
        }
    }

    @PutMapping("/{id}")
    public ApiResponse<CourseDTO> updateCourse(
            @PathVariable String id, 
            @RequestBody UpdateCourseRequest request) {
        try {
            CourseDTO course = courseService.updateCourse(id, request);
            return ApiResponse.success(course);
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.INVALID_PARAM, e.getMessage());
        }
    }

    @DeleteMapping("/{id}")
    public ApiResponse<Void> deleteCourse(@PathVariable String id) {
        try {
            courseService.deleteCourse(id);
            return ApiResponse.success();
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.COURSE_NOT_FOUND);
        }
    }
}
