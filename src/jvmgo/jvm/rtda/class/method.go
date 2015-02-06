package class

import (
    "fmt"
    . "jvmgo/any"
    cf "jvmgo/classfile"
)

const (
    mainMethodName              = "main"
    mainMethodDesc              = "([Ljava/lang/String;)V"
    clinitMethodName            = "<clinit>"
    clinitMethodDesc            = "()V"
    constructorName             = "<init>"
    registerNativesMethodName   = "registerNatives"
    registerNativesMethodDesc   = "()V"
)

type Method struct {
    ClassMember
    ExceptionTable
    maxStack        uint
    maxLocals       uint
    argCount        uint
    code            []byte
    nativeMethod    Any // cannot use package 'native' because of cycle import!
    parameterTypes  []*Class
}

func newMethod(class *Class, methodInfo *cf.MethodInfo) (*Method) {
    method := &Method{}
    method.class = class
    method.accessFlags = methodInfo.AccessFlags()
    method.name = methodInfo.Name()
    method.descriptor = methodInfo.Descriptor()
    method.argCount = calcArgCount(method.descriptor)
    if codeAttr := methodInfo.CodeAttribute(); codeAttr != nil {
        method.code = codeAttr.Code()
        method.maxStack = codeAttr.MaxStack()
        method.maxLocals = codeAttr.MaxLocals()
        rtCp := method.class.constantPool
        if len(codeAttr.ExceptionTable()) > 0 {
            method.copyExceptionTable(codeAttr.ExceptionTable(), rtCp)
        }
    }
    
    return method
}

func (self *Method) String() string {
    return fmt.Sprintf("{Method name:%v descriptor:%v}", self.name, self.descriptor)
}

// getters
func (self *Method) MaxStack() (uint) {
    return self.maxStack
}
func (self *Method) MaxLocals() (uint) {
    return self.maxLocals
}
func (self *Method) ArgCount() (uint) {
    return self.argCount
}
func (self *Method) Code() ([]byte) {
    return self.code
}
func (self *Method) NativeMethod() (Any) {
    return self.nativeMethod
}

// argCount for static method
// argCount+1 for instance method
func (self *Method) ActualArgCount() (uint) {
    if self.IsStatic() {
        return self.argCount
    } else {
        return self.argCount + 1
    }
}

func (self *Method) IsVoidReturnType() bool {
    return isVoidReturnType(self.descriptor)
}

func (self *Method) isConstructor() (bool) {
    return !self.IsStatic() && self.name == constructorName
}
func (self *Method) IsClinit() (bool) {
    return self.IsStatic() &&
            self.name == clinitMethodName && 
            self.descriptor == clinitMethodDesc
}
func (self *Method) IsRegisterNatives() (bool) {
    return self.IsStatic() &&
            self.name == registerNativesMethodName &&
            self.descriptor == registerNativesMethodDesc

}

// reflection
func (self *Method) ParameterTypes() ([]*Class) {
    if self.parameterTypes == nil {
        self.resolveParameterTypes()
    }
    return self.parameterTypes
}
func (self *Method) resolveParameterTypes() {
    if self.argCount == 0 {
        // todo optimize
        self.parameterTypes = make([]*Class, 0)
    }

    // todo
    md := parseMethodDescriptor(self.descriptor)
    self.parameterTypes = make([]*Class, len(md.parameterTypes))
    for i, paramType := range md.parameterTypes {
        self.parameterTypes[i] = self.class.classLoader.LoadClass(paramType.descriptor) // todo
    }
}