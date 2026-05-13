package com.example.converter.controller;

import com.example.converter.dto.ConvertRequest;
import com.example.converter.dto.ConvertResponse;
import com.example.converter.service.ConversionService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/convert")
@RequiredArgsConstructor
public class ConvertController {

    private final ConversionService conversionService;

    @PostMapping("/xml-to-json")
    public ResponseEntity<ConvertResponse> xmlToJson(@RequestBody ConvertRequest request) throws Exception {
        String result = conversionService.convertXmlToJson(
                request.getData(),
                request.getMappingName(),
                Boolean.TRUE.equals(request.getValidate())
        );

        ConvertResponse response = new ConvertResponse();
        response.setResult(result);
        response.setFormat("json");
        response.setValidated(Boolean.TRUE.equals(request.getValidate()));

        return ResponseEntity.ok()
                .contentType(MediaType.APPLICATION_JSON)
                .body(response);
    }

    @PostMapping(value = "/json-to-xml")
    public ResponseEntity<ConvertResponse> jsonToXml(@RequestBody ConvertRequest request) throws Exception {
        String result = conversionService.convertJsonToXml(
                request.getData(),
                request.getMappingName(),
                Boolean.TRUE.equals(request.getValidate())
        );

        ConvertResponse response = new ConvertResponse();
        response.setResult(result);
        response.setFormat("xml");
        response.setValidated(Boolean.TRUE.equals(request.getValidate()));

        return ResponseEntity.ok()
                .contentType(MediaType.APPLICATION_XML)
                .body(response);
    }
}
