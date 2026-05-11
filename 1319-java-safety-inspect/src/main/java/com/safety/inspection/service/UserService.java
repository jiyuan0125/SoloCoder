package com.safety.inspection.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.safety.inspection.common.BusinessException;
import com.safety.inspection.dto.LoginDTO;
import com.safety.inspection.dto.LoginVO;
import com.safety.inspection.dto.UserDTO;
import com.safety.inspection.entity.Role;
import com.safety.inspection.entity.User;
import com.safety.inspection.entity.UserRole;
import com.safety.inspection.enums.RoleEnum;
import com.safety.inspection.mapper.RoleMapper;
import com.safety.inspection.mapper.UserMapper;
import com.safety.inspection.mapper.UserRoleMapper;
import com.safety.inspection.security.JwtTokenProvider;
import lombok.RequiredArgsConstructor;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.StringUtils;

import java.util.List;

@Service
@RequiredArgsConstructor
public class UserService extends ServiceImpl<UserMapper, User> {

    private final UserMapper userMapper;
    private final RoleMapper roleMapper;
    private final UserRoleMapper userRoleMapper;
    private final PasswordEncoder passwordEncoder;
    private final JwtTokenProvider jwtTokenProvider;

    public LoginVO login(LoginDTO loginDTO) {
        User user = userMapper.selectOne(
            new LambdaQueryWrapper<User>()
                .eq(User::getUsername, loginDTO.getUsername())
                .eq(User::getStatus, 1)
        );

        if (user == null) {
            throw new BusinessException("用户不存在或已禁用");
        }

        if (!passwordEncoder.matches(loginDTO.getPassword(), user.getPassword())) {
            throw new BusinessException("密码错误");
        }

        List<String> roles = userMapper.selectRoleCodesByUserId(user.getId());
        String role = roles.isEmpty() ? RoleEnum.INSPECTOR.getCode() : roles.get(0);

        String token = jwtTokenProvider.generateToken(user.getId(), user.getUsername(), role);

        LoginVO vo = new LoginVO();
        vo.setUserId(user.getId());
        vo.setUsername(user.getUsername());
        vo.setRealName(user.getRealName());
        vo.setToken(token);
        vo.setRole(role);

        return vo;
    }

    @Transactional
    public void createUser(UserDTO userDTO) {
        User existUser = userMapper.selectOne(
            new LambdaQueryWrapper<User>().eq(User::getUsername, userDTO.getUsername())
        );

        if (existUser != null) {
            throw new BusinessException("用户名已存在");
        }

        User user = new User();
        user.setUsername(userDTO.getUsername());
        user.setPassword(passwordEncoder.encode(userDTO.getPassword()));
        user.setRealName(userDTO.getRealName());
        user.setPhone(userDTO.getPhone());
        user.setEmail(userDTO.getEmail());
        user.setDepartmentId(userDTO.getDepartmentId());
        user.setStatus(userDTO.getStatus() != null ? userDTO.getStatus() : 1);
        userMapper.insert(user);

        if (userDTO.getRoleIds() != null && !userDTO.getRoleIds().isEmpty()) {
            for (Long roleId : userDTO.getRoleIds()) {
                UserRole userRole = new UserRole();
                userRole.setUserId(user.getId());
                userRole.setRoleId(roleId);
                userRoleMapper.insert(userRole);
            }
        }
    }

    @Transactional
    public void updateUser(UserDTO userDTO) {
        User user = userMapper.selectById(userDTO.getId());
        if (user == null) {
            throw new BusinessException("用户不存在");
        }

        if (StringUtils.hasText(userDTO.getPassword())) {
            user.setPassword(passwordEncoder.encode(userDTO.getPassword()));
        }
        user.setRealName(userDTO.getRealName());
        user.setPhone(userDTO.getPhone());
        user.setEmail(userDTO.getEmail());
        user.setDepartmentId(userDTO.getDepartmentId());
        user.setStatus(userDTO.getStatus());
        userMapper.updateById(user);

        if (userDTO.getRoleIds() != null) {
            userRoleMapper.deleteByUserId(user.getId());
            for (Long roleId : userDTO.getRoleIds()) {
                UserRole userRole = new UserRole();
                userRole.setUserId(user.getId());
                userRole.setRoleId(roleId);
                userRoleMapper.insert(userRole);
            }
        }
    }

    public void deleteUser(Long id) {
        User user = userMapper.selectById(id);
        if (user == null) {
            throw new BusinessException("用户不存在");
        }
        userMapper.deleteById(id);
        userRoleMapper.deleteByUserId(id);
    }

    public Page<User> getUserPage(int pageNum, int pageSize, String keyword, Long departmentId) {
        Page<User> page = new Page<>(pageNum, pageSize);
        LambdaQueryWrapper<User> wrapper = new LambdaQueryWrapper<>();

        if (StringUtils.hasText(keyword)) {
            wrapper.and(w -> w
                .like(User::getUsername, keyword)
                .or()
                .like(User::getRealName, keyword)
            );
        }
        if (departmentId != null) {
            wrapper.eq(User::getDepartmentId, departmentId);
        }
        wrapper.orderByDesc(User::getCreatedAt);

        return userMapper.selectPage(page, wrapper);
    }

    public List<Role> getAllRoles() {
        return roleMapper.selectList(new LambdaQueryWrapper<Role>().orderByAsc(Role::getId));
    }

    public List<User> getUsersByRole(String roleCode) {
        return userMapper.selectUsersByRoleCode(roleCode);
    }

    public User getUserById(Long id) {
        return userMapper.selectById(id);
    }

    public List<Long> getUserRoleIds(Long userId) {
        return userRoleMapper.selectRoleIdsByUserId(userId);
    }
}
