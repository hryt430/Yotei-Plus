"use client"

import type React from "react"
import { useState } from "react"
import { Button } from "@/components/ui/forms/button"
import { Input } from "@/components/ui/forms/input"
import { Label } from "@/components/ui/forms/label"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/data-display/card"
import { Checkbox } from "@/components/ui/forms/checkbox"
import { Eye, EyeOff, Mail, Lock, User, Building, ArrowRight, Check, Loader2 } from "lucide-react"
import Link from "next/link"
import { useAuth } from "@/providers/auth-provider"
import { isValidEmail, isValidPassword, isValidUsername } from "@/lib/utils"

export default function RegisterForm() {
  const { register, isLoading, error, clearError } = useAuth()
  
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const [formData, setFormData] = useState({
    firstName: "",
    lastName: "",
    email: "",
    company: "",
    password: "",
    confirmPassword: "",
    agreeToTerms: false,
    subscribeNewsletter: false,
  })

  const [formErrors, setFormErrors] = useState<{
    firstName?: string
    lastName?: string
    email?: string
    password?: string
    confirmPassword?: string
    agreeToTerms?: string
  }>({})

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type, checked } = e.target
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value,
    }))
    
    // エラーをクリア
    if (formErrors[name as keyof typeof formErrors]) {
      setFormErrors(prev => ({
        ...prev,
        [name]: undefined,
      }))
    }
    
    // グローバルエラーもクリア
    if (error) {
      clearError()
    }
  }

  const validateForm = (): boolean => {
    const errors: typeof formErrors = {}
    
    if (!formData.firstName.trim()) {
      errors.firstName = '名前を入力してください'
    }
    
    if (!formData.lastName.trim()) {
      errors.lastName = '姓を入力してください'
    }
    
    if (!formData.email) {
      errors.email = 'メールアドレスを入力してください'
    } else if (!isValidEmail(formData.email)) {
      errors.email = '有効なメールアドレスを入力してください'
    }
    
    if (!formData.password) {
      errors.password = 'パスワードを入力してください'
    } else if (!isValidPassword(formData.password)) {
      errors.password = 'パスワードは8文字以上で入力してください'
    }
    
    if (!formData.confirmPassword) {
      errors.confirmPassword = 'パスワード確認を入力してください'
    } else if (formData.password !== formData.confirmPassword) {
      errors.confirmPassword = 'パスワードが一致しません'
    }
    
    if (!formData.agreeToTerms) {
      errors.agreeToTerms = '利用規約に同意してください'
    }
    
    setFormErrors(errors)
    return Object.keys(errors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!validateForm()) {
      return
    }
    
    try {
      // usernameとしてfirstName + lastNameを使用
      const username = `${formData.firstName.trim()} ${formData.lastName.trim()}`
      await register(username, formData.email, formData.password)
    } catch (error) {
      // エラーはAuthProviderで処理される
      console.error('Registration failed:', error)
    }
  }

  const passwordRequirements = [
    { text: "8文字以上", met: formData.password.length >= 8 },
    { text: "大文字を含む", met: /[A-Z]/.test(formData.password) },
    { text: "小文字を含む", met: /[a-z]/.test(formData.password) },
    { text: "数字を含む", met: /\d/.test(formData.password) },
  ]

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 flex items-center justify-center p-4">
      <div className="w-full max-w-lg">
        {/* Logo/Brand Section */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-gray-900 rounded-xl mb-4 shadow-lg">
            <div className="text-white font-bold text-xl">TF</div>
          </div>
          <h1 className="text-2xl font-bold text-gray-900 mb-2">アカウントを作成</h1>
          <p className="text-gray-600">効率的にタスクを管理しましょう</p>
        </div>

        {/* Registration Form */}
        <Card className="shadow-xl border-0 bg-white/80 backdrop-blur-sm">
          <CardHeader className="space-y-1 pb-4">
            <CardTitle className="text-xl font-semibold text-center">サインアップ</CardTitle>
            <CardDescription className="text-center">TaskFlowを信頼する数千のユーザーに参加</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-4">
              {/* Name Fields */}
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-2">
                  <Label htmlFor="firstName" className="text-sm font-medium text-gray-700">
                    名前
                  </Label>
                  <div className="relative">
                    <User className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
                    <Input
                      id="firstName"
                      name="firstName"
                      value={formData.firstName}
                      onChange={handleChange}
                      placeholder="太郎"
                      className={`pl-10 border-gray-200 focus:border-gray-400 focus:ring-gray-400 ${
                        formErrors.firstName ? 'border-red-300 focus:border-red-400 focus:ring-red-400' : ''
                      }`}
                      required
                    />
                  </div>
                  {formErrors.firstName && (
                    <p className="text-xs text-red-600">{formErrors.firstName}</p>
                  )}
                </div>
                <div className="space-y-2">
                  <Label htmlFor="lastName" className="text-sm font-medium text-gray-700">
                    姓
                  </Label>
                  <Input
                    id="lastName"
                    name="lastName"
                    value={formData.lastName}
                    onChange={handleChange}
                    placeholder="田中"
                    className={`border-gray-200 focus:border-gray-400 focus:ring-gray-400 ${
                      formErrors.lastName ? 'border-red-300 focus:border-red-400 focus:ring-red-400' : ''
                    }`}
                    required
                  />
                  {formErrors.lastName && (
                    <p className="text-xs text-red-600">{formErrors.lastName}</p>
                  )}
                </div>
              </div>

              {/* Email Field */}
              <div className="space-y-2">
                <Label htmlFor="email" className="text-sm font-medium text-gray-700">
                  メールアドレス
                </Label>
                <div className="relative">
                  <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
                  <Input
                    id="email"
                    name="email"
                    type="email"
                    value={formData.email}
                    onChange={handleChange}
                    placeholder="taro@example.com"
                    className={`pl-10 border-gray-200 focus:border-gray-400 focus:ring-gray-400 ${
                      formErrors.email ? 'border-red-300 focus:border-red-400 focus:ring-red-400' : ''
                    }`}
                    required
                  />
                </div>
                {formErrors.email && (
                  <p className="text-sm text-red-600">{formErrors.email}</p>
                )}
              </div>

              {/* Company Field */}
              <div className="space-y-2">
                <Label htmlFor="company" className="text-sm font-medium text-gray-700">
                  会社名（任意）
                </Label>
                <div className="relative">
                  <Building className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
                  <Input
                    id="company"
                    name="company"
                    value={formData.company}
                    onChange={handleChange}
                    placeholder="あなたの会社名"
                    className="pl-10 border-gray-200 focus:border-gray-400 focus:ring-gray-400"
                  />
                </div>
              </div>

              {/* Password Field */}
              <div className="space-y-2">
                <Label htmlFor="password" className="text-sm font-medium text-gray-700">
                  パスワード
                </Label>
                <div className="relative">
                  <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
                  <Input
                    id="password"
                    name="password"
                    type={showPassword ? "text" : "password"}
                    value={formData.password}
                    onChange={handleChange}
                    placeholder="強力なパスワードを作成"
                    className={`pl-10 pr-10 border-gray-200 focus:border-gray-400 focus:ring-gray-400 ${
                      formErrors.password ? 'border-red-300 focus:border-red-400 focus:ring-red-400' : ''
                    }`}
                    required
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600 transition-colors"
                  >
                    {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </button>
                </div>
                {formErrors.password && (
                  <p className="text-sm text-red-600">{formErrors.password}</p>
                )}
                {/* Password Requirements */}
                {formData.password && (
                  <div className="mt-2 space-y-1">
                    {passwordRequirements.map((req, index) => (
                      <div key={index} className="flex items-center text-xs">
                        <Check className={`w-3 h-3 mr-2 ${req.met ? "text-green-500" : "text-gray-300"}`} />
                        <span className={req.met ? "text-green-600" : "text-gray-500"}>{req.text}</span>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {/* Confirm Password Field */}
              <div className="space-y-2">
                <Label htmlFor="confirmPassword" className="text-sm font-medium text-gray-700">
                  パスワード確認
                </Label>
                <div className="relative">
                  <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
                  <Input
                    id="confirmPassword"
                    name="confirmPassword"
                    type={showConfirmPassword ? "text" : "password"}
                    value={formData.confirmPassword}
                    onChange={handleChange}
                    placeholder="パスワードを再入力"
                    className={`pl-10 pr-10 border-gray-200 focus:border-gray-400 focus:ring-gray-400 ${
                      formErrors.confirmPassword ? 'border-red-300 focus:border-red-400 focus:ring-red-400' : ''
                    }`}
                    required
                  />
                  <button
                    type="button"
                    onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                    className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600 transition-colors"
                  >
                    {showConfirmPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </button>
                </div>
                {formErrors.confirmPassword && (
                  <p className="text-sm text-red-600">{formErrors.confirmPassword}</p>
                )}
                {formData.confirmPassword && formData.password !== formData.confirmPassword && !formErrors.confirmPassword && (
                  <p className="text-xs text-red-500">パスワードが一致しません</p>
                )}
              </div>

              {/* Global Error Message */}
              {error && (
                <div className="rounded-md bg-red-50 p-4 border border-red-200">
                  <div className="text-sm text-red-700">{error}</div>
                </div>
              )}

              {/* Checkboxes */}
              <div className="space-y-3">
                <div className="flex items-start space-x-2">
                  <Checkbox
                    id="agreeToTerms"
                    name="agreeToTerms"
                    checked={formData.agreeToTerms}
                    onCheckedChange={(checked) => setFormData({ ...formData, agreeToTerms: !!checked })}
                    className="mt-0.5"
                  />
                  <Label htmlFor="agreeToTerms" className="text-sm text-gray-600 leading-relaxed">
                    <Link href="/terms" className="text-gray-900 hover:text-gray-700 transition-colors">
                      利用規約
                    </Link>{" "}
                    および{" "}
                    <Link href="/privacy" className="text-gray-900 hover:text-gray-700 transition-colors">
                      プライバシーポリシー
                    </Link>
                    に同意します
                  </Label>
                </div>
                {formErrors.agreeToTerms && (
                  <p className="text-sm text-red-600">{formErrors.agreeToTerms}</p>
                )}
                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="subscribeNewsletter"
                    name="subscribeNewsletter"
                    checked={formData.subscribeNewsletter}
                    onCheckedChange={(checked) => setFormData({ ...formData, subscribeNewsletter: !!checked })}
                  />
                  <Label htmlFor="subscribeNewsletter" className="text-sm text-gray-600">
                    製品のアップデートとヒントを受け取る
                  </Label>
                </div>
              </div>

              {/* Submit Button */}
              <Button
                type="submit"
                disabled={isLoading || !formData.agreeToTerms || formData.password !== formData.confirmPassword}
                className="w-full bg-gray-900 hover:bg-gray-800 text-white shadow-lg hover:shadow-xl transition-all duration-300 hover:-translate-y-0.5 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:translate-y-0"
              >
                {isLoading ? (
                  <>
                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                    アカウント作成中...
                  </>
                ) : (
                  <>
                    アカウント作成
                    <ArrowRight className="w-4 h-4 ml-2" />
                  </>
                )}
              </Button>
            </form>

            {/* Divider */}
            <div className="relative my-6">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-gray-200" />
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-2 bg-white text-gray-500">または</span>
              </div>
            </div>

            {/* Sign In Link */}
            <div className="text-center">
              <p className="text-sm text-gray-600">
                すでにアカウントをお持ちですか？{" "}
                <Link href="/auth/login" className="font-medium text-gray-900 hover:text-gray-700 transition-colors">
                  サインイン
                </Link>
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}