package com.edi.converter.controller;

import com.edi.converter.dto.GenerateRequest;
import com.edi.converter.dto.ParseRequest;
import com.edi.converter.model.EdiDocument;
import com.edi.converter.service.EdiConverterService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/convert")
public class ConvertController {

    private final EdiConverterService converterService;

    @Autowired
    public ConvertController(EdiConverterService converterService) {
        this.converterService = converterService;
    }

    @PostMapping("/parse")
    public ResponseEntity<EdiDocument> parse(@RequestBody ParseRequest request) {
        EdiDocument result = converterService.parse(request.getEdiText(), request.getPartnerId());
        return ResponseEntity.ok(result);
    }

    @PostMapping("/generate")
    public ResponseEntity<String> generate(@RequestBody GenerateRequest request) {
        String result = converterService.generate(request.getDocument(), request.getPartnerId());
        return ResponseEntity.ok()
                .contentType(org.springframework.http.MediaType.TEXT_PLAIN)
                .body(result);
    }
}
