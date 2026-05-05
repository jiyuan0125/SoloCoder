package com.workshift.server.controller;

import com.workshift.common.dto.ShiftSwapDTO;
import com.workshift.common.enums.ErrorCode;
import com.workshift.common.request.ConfirmSwapRequest;
import com.workshift.common.request.CreateSwapRequest;
import com.workshift.common.response.ApiResponse;
import com.workshift.server.service.SwapService;
import java.util.List;
import java.util.Optional;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/swaps")
public class SwapController {

    private final SwapService swapService;

    public SwapController(SwapService swapService) {
        this.swapService = swapService;
    }

    @PostMapping
    public ApiResponse<ShiftSwapDTO> createSwap(@RequestBody CreateSwapRequest request) {
        if (request.getRequesterEmployeeId() == null || 
            request.getTargetEmployeeId() == null ||
            request.getRequesterDate() == null ||
            request.getTargetDate() == null) {
            return ApiResponse.error(ErrorCode.INVALID_PARAMETER.getCode(), "参数不完整");
        }

        Optional<ErrorCode> result = swapService.createSwap(
                request.getRequesterEmployeeId(),
                request.getTargetEmployeeId(),
                request.getRequesterDate(),
                request.getTargetDate(),
                request.getRequesterSignature());

        if (result.isPresent()) {
            return ApiResponse.error(result.get().getCode(), result.get().getMessage());
        }

        List<ShiftSwapDTO> swaps = swapService.getSwapsByEmployee(request.getRequesterEmployeeId());
        ShiftSwapDTO latest = swaps.get(swaps.size() - 1);
        return ApiResponse.success(latest);
    }

    @PutMapping("/confirm")
    public ApiResponse<Void> confirmSwap(@RequestBody ConfirmSwapRequest request) {
        if (request.getSwapId() == null || request.getTargetEmployeeId() == null) {
            return ApiResponse.error(ErrorCode.INVALID_PARAMETER.getCode(), "参数不完整");
        }

        Optional<ErrorCode> result = swapService.confirmSwap(
                request.getSwapId(),
                request.getTargetEmployeeId(),
                request.isConfirmed(),
                request.getTargetSignature(),
                request.getRejectReason());

        if (result.isPresent()) {
            return ApiResponse.error(result.get().getCode(), result.get().getMessage());
        }
        return ApiResponse.success();
    }

    @GetMapping("/{id}")
    public ApiResponse<ShiftSwapDTO> getSwap(@PathVariable String id) {
        ShiftSwapDTO swap = swapService.getSwapById(id);
        if (swap == null) {
            return ApiResponse.error(ErrorCode.SWAP_NOT_FOUND.getCode(), ErrorCode.SWAP_NOT_FOUND.getMessage());
        }
        return ApiResponse.success(swap);
    }

    @GetMapping("/employee/{employeeId}")
    public ApiResponse<List<ShiftSwapDTO>> getSwapsByEmployee(@PathVariable String employeeId) {
        List<ShiftSwapDTO> swaps = swapService.getSwapsByEmployee(employeeId);
        return ApiResponse.success(swaps);
    }

    @GetMapping
    public ApiResponse<List<ShiftSwapDTO>> getAllSwaps() {
        List<ShiftSwapDTO> swaps = swapService.getAllSwaps();
        return ApiResponse.success(swaps);
    }
}